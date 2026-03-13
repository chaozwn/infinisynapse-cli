package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/types"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI conversation and configuration",
}

var aiChatCmd = &cobra.Command{
	Use:   "chat [message]",
	Short: "Start a new AI conversation or continue an existing one",
	Long: `Send a message to the AI and receive streaming responses.
Use --task-id to continue an existing conversation.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		taskID, _ := cmd.Flags().GetString("task-id")
		connID := uuid.New().String()

		msgType := "newTask"
		if taskID != "" {
			msgType = "askResponse"
		}

		msg := types.WebviewMessage{
			Type:   msgType,
			Text:   args[0],
			ConnID: connID,
			TaskID: taskID,
		}
		if msgType == "askResponse" {
			msg.AskResponse = "messageResponse"
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		sseClient := client.NewSSEClient(c)
		sseDone := make(chan error, 1)
		sseReady := make(chan struct{})
		textLenByTs := make(map[int64]int)

		skipSayTypes := map[string]bool{
			"task":             true,
			"user_feedback":    true,
			"api_req_started":  true,
			"api_req_finished": true,
			"api_req_retried":  true,
			"deleted_api_reqs": true,
		}

		printableSayTypes := map[string]bool{
			"text":              true,
			"completion_result": true,
			"reasoning":         true,
			"error":             true,
		}

		go func() {
			err := sseClient.Subscribe(ctx, fmt.Sprintf("/api/ai/events?connId=%s", connID), sseReady, func(event client.SSEEvent) bool {
				if event.Event == "heartbeat" {
					return true
				}

				var payload struct {
					TaskID  string `json:"taskId"`
					Message struct {
						Ts        int64  `json:"ts"`
						Type      string `json:"type"`
						Say       string `json:"say"`
						Ask       string `json:"ask"`
						Text      string `json:"text"`
						Reasoning string `json:"reasoning"`
						Partial   bool   `json:"partial"`
					} `json:"message"`
				}

			switch event.Event {
				case "message.partial":
					if err := json.Unmarshal([]byte(event.Data), &payload); err != nil {
						return true
					}
					msg := payload.Message

					if msg.Type == "ask" && !msg.Partial {
						return false
					}

					if msg.Type != "say" || skipSayTypes[msg.Say] {
						return true
					}
					if !printableSayTypes[msg.Say] {
						return true
					}

					text := msg.Text
					if msg.Say == "reasoning" && msg.Reasoning != "" {
						text = msg.Reasoning
					}

					lastLen := textLenByTs[msg.Ts]
					if text != "" && len(text) > lastLen {
						delta := text[lastLen:]
						fmt.Print(delta)
						textLenByTs[msg.Ts] = len(text)
					}

				case "message.add":
					if err := json.Unmarshal([]byte(event.Data), &payload); err != nil {
						return true
					}
					msg := payload.Message

					if msg.Type == "say" && printableSayTypes[msg.Say] {
						text := msg.Text
						lastLen := textLenByTs[msg.Ts]
						if text != "" && len(text) > lastLen {
							fmt.Print(text[lastLen:])
						}
					}

					if msg.Type == "ask" && !msg.Partial {
						return false
					}
					if msg.Say == "completion_result" {
						return false
					}

				case "message.update":
					return true

				case "state.ready":
					return true

				case "notification":
					var notif struct {
						Type    string `json:"type"`
						Title   string `json:"title"`
						Message string `json:"message"`
					}
					if err := json.Unmarshal([]byte(event.Data), &notif); err == nil {
						if notif.Type == "error" {
							fmt.Fprintf(os.Stderr, "\n[%s] %s: %s\n", notif.Type, notif.Title, notif.Message)
						}
					}
					return true
				}

				return true
			})
			sseDone <- err
		}()

		select {
		case <-sseReady:
		case err := <-sseDone:
			return fmt.Errorf("SSE connection failed before sending message: %w", err)
		case <-time.After(10 * time.Second):
			cancel()
			return fmt.Errorf("SSE connection timeout (10s), server may be unreachable")
		}

		postDone := make(chan error, 1)
		go func() {
			_, err := c.Post("/api/ai/message", msg)
			postDone <- err
		}()

		select {
		case sseErr := <-sseDone:
			if sseErr != nil && sseErr != context.Canceled {
				output.PrintError("SSE stream error: %v", sseErr)
			}
		case <-ctx.Done():
		}

		fmt.Println()
		return nil
	},
}

var aiStateCmd = &cobra.Command{
	Use:   "state",
	Short: "Get AI state",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		taskID, _ := cmd.Flags().GetString("task-id")
		params := map[string]string{}
		if taskID != "" {
			params["taskId"] = taskID
		}

		data, err := c.Get("/api/ai/state", params)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var aiConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Get or update API configuration",
}

var aiConfigGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current API configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Get("/api/ai/configuration", nil)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var aiConfigSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Update API configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		provider, _ := cmd.Flags().GetString("provider")
		modelID, _ := cmd.Flags().GetString("model")
		apiKey, _ := cmd.Flags().GetString("api-key")
		baseURL, _ := cmd.Flags().GetString("base-url")

		apiConfig := &types.APIConfiguration{}
		if provider != "" {
			apiConfig.APIProvider = provider
		}
		if modelID != "" {
			apiConfig.OpenAIModelID = modelID
		}
		if apiKey != "" {
			apiConfig.OpenAIAPIKey = apiKey
		}
		if baseURL != "" {
			apiConfig.OpenAIBaseURL = baseURL
		}

		update := types.AISettingsUpdate{
			APIConfiguration: apiConfig,
		}

		data, err := c.Post("/api/ai/settings", update)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var aiModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available AI models",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Get("/api/ai/models", nil)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var aiCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a running task",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		taskID, _ := cmd.Flags().GetString("task-id")
		if taskID == "" {
			return fmt.Errorf("--task-id is required")
		}

		msg := types.WebviewMessage{
			Type:   "cancelTask",
			TaskID: taskID,
		}

		data, err := c.Post("/api/ai/message", msg)
		if err != nil {
			return err
		}

		output.PrintSuccess("Task %s cancelled", taskID)

		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

func init() {
	aiChatCmd.Flags().String("task-id", "", "Continue conversation in existing task")

	aiStateCmd.Flags().String("task-id", "", "Get state for specific task")

	aiConfigSetCmd.Flags().String("provider", "", "API provider (openai, anthropic, deepseek, qwen, infinisynapse)")
	aiConfigSetCmd.Flags().String("model", "", "Model ID")
	aiConfigSetCmd.Flags().String("api-key", "", "API key")
	aiConfigSetCmd.Flags().String("base-url", "", "API base URL")

	aiCancelCmd.Flags().String("task-id", "", "Task ID to cancel (required)")

	aiConfigCmd.AddCommand(aiConfigGetCmd)
	aiConfigCmd.AddCommand(aiConfigSetCmd)

	aiCmd.AddCommand(aiChatCmd)
	aiCmd.AddCommand(aiStateCmd)
	aiCmd.AddCommand(aiConfigCmd)
	aiCmd.AddCommand(aiModelsCmd)
	aiCmd.AddCommand(aiCancelCmd)

	rootCmd.AddCommand(aiCmd)
}
