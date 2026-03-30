package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/session"
	"github.com/chaozwn/infinisynapse-cli/internal/task"
	"github.com/chaozwn/infinisynapse-cli/internal/types"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions and chat",
	Long: `Manage sessions for multi-turn chat conversations.

Create a new session and start a task:
  agent_infini session new --name "analysis" --query "Analyze sales data"

Continue a conversation in an existing session:
  agent_infini session -s "analysis" --query "Focus on revenue"

Manage sessions:
  agent_infini session ls
  agent_infini session show "analysis"
  agent_infini session rm "analysis"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query, _ := cmd.Flags().GetString("query")
		if query == "" {
			return cmd.Help()
		}
		if sessionName == "" {
			return fmt.Errorf("--session / -s is required when using --query")
		}

		sess, err := session.Load(sessionName)
		if err != nil {
			return fmt.Errorf("session '%s' not found: %w", sessionName, err)
		}
		if sess.TaskID == "" {
			return fmt.Errorf("session '%s' has no active task; use 'session new' to start one", sessionName)
		}

		result, err := task.RunAskResponse(sess.TaskID, query)
		if err != nil {
			return err
		}
		if result.TaskID != "" {
			saveSession(sessionName, result.TaskID, result.ConnID, result.Status, result.LastAskType)
		}
		return nil
	},
}

var sessionNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new session and start a task",
	Long: `Create a new session and send a newTask to the server.

If a session with the same name already exists, it will be reset with a new task.

Examples:
  agent_infini session new --name "data-analysis" --query "Analyze sales data"
  agent_infini session new --name "inventory" --query "Check stock levels"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		query, _ := cmd.Flags().GetString("query")
		if name == "" || query == "" {
			return fmt.Errorf("both --name and --query are required")
		}

		result, err := task.RunNewTask(query)
		if err != nil {
			return err
		}
		if result.TaskID != "" {
			saveSession(name, result.TaskID, result.ConnID, result.Status, result.LastAskType)
		}
		return nil
	},
}

var sessionListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			output.PrintSuccess("No sessions found.")
			return nil
		}

		printer := output.NewPrinter(getOutputFormat())
		headers := []string{"Name", "Task ID", "Status", "Updated At"}
		rows := make([][]string, 0, len(sessions))
		for _, s := range sessions {
			taskID := s.TaskID
			if len(taskID) > 16 {
				taskID = taskID[:16] + "..."
			}
			status := s.Status
			if status == "" {
				status = "-"
			}
			rows = append(rows, []string{s.Name, taskID, status, s.UpdatedAt.Format("2006-01-02 15:04:05")})
		}
		printer.PrintTable(headers, rows)
		return nil
	},
}

var sessionShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show session details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := session.Load(args[0])
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(sess)
	},
}

var sessionRemoveCmd = &cobra.Command{
	Use:     "rm [name]",
	Aliases: []string{"delete", "remove"},
	Short:   "Delete a session",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := session.Delete(args[0]); err != nil {
			return err
		}
		output.PrintSuccess("Session '%s' deleted", args[0])
		return nil
	},
}

var sessionStateCmd = &cobra.Command{
	Use:   "state",
	Short: "Get AI state for a session",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides("", "")
		if err != nil {
			return err
		}

		taskID, _ := cmd.Flags().GetString("task-id")
		if taskID == "" && sessionName != "" {
			sess, loadErr := session.Load(sessionName)
			if loadErr == nil {
				taskID = sess.TaskID
			}
		}

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

var sessionCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a running task in a session",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides("", "")
		if err != nil {
			return err
		}

		taskID, _ := cmd.Flags().GetString("task-id")
		if taskID == "" && sessionName != "" {
			sess, loadErr := session.Load(sessionName)
			if loadErr == nil {
				taskID = sess.TaskID
			}
		}
		if taskID == "" {
			return fmt.Errorf("--task-id or --session is required")
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

		if sessionName != "" {
			saveSession(sessionName, taskID, "", "cancelled", "")
		}

		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

func init() {
	sessionCmd.Flags().StringP("query", "q", "", "Send a message to continue the conversation (askResponse)")

	sessionNewCmd.Flags().String("name", "", "Session name (required)")
	sessionNewCmd.Flags().String("query", "", "Initial message/query (required)")
	_ = sessionNewCmd.MarkFlagRequired("name")
	_ = sessionNewCmd.MarkFlagRequired("query")

	sessionStateCmd.Flags().String("task-id", "", "Get state for specific task")
	sessionCancelCmd.Flags().String("task-id", "", "Task ID to cancel")

	sessionCmd.AddCommand(sessionNewCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionShowCmd)
	sessionCmd.AddCommand(sessionRemoveCmd)
	sessionCmd.AddCommand(sessionStateCmd)
	sessionCmd.AddCommand(sessionCancelCmd)

	rootCmd.AddCommand(sessionCmd)
}

func saveSession(name, taskID, connID, status, lastAskType string) {
	if name == "" || taskID == "" {
		return
	}
	userID := config.GetUserID()
	wsPath := taskWorkspacePath(userID, taskID)
	_ = session.Save(name, userID, taskID, connID, wsPath, status, lastAskType)
}

func taskWorkspacePath(userID, taskID string) string {
	if userID == "" {
		cwd, _ := os.Getwd()
		return cwd
	}
	home, err := os.UserHomeDir()
	if err != nil {
		cwd, _ := os.Getwd()
		return cwd
	}
	return filepath.Join(home, ".infiniSynapse", "tasks", userID, taskID)
}
