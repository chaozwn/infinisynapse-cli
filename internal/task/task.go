package task

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
)

type StreamResult struct {
	TaskID      string
	ConnID      string
	Status      string
	LastAskType string
}

func RunNewTask(query string, jsonMode bool) (*StreamResult, error) {
	connID := uuid.New().String()
	msg := types.WebviewMessage{
		Type:   "newTask",
		Text:   query,
		ConnID: connID,
	}
	return runStreamingChat(connID, msg, jsonMode)
}

func RunAskResponse(taskID, query string, jsonMode bool) (*StreamResult, error) {
	connID := uuid.New().String()
	msg := types.WebviewMessage{
		Type:        "askResponse",
		Text:        query,
		ConnID:      connID,
		TaskID:      taskID,
		AskResponse: "messageResponse",
	}
	return runStreamingChat(connID, msg, jsonMode)
}

type StreamEvent struct {
	TaskID    string `json:"taskId"`
	Ts        int64  `json:"ts"`
	Type      string `json:"type"`
	Say       string `json:"say,omitempty"`
	Ask       string `json:"ask,omitempty"`
	Text      string `json:"text,omitempty"`
	Reasoning string `json:"reasoning,omitempty"`
	Partial   bool   `json:"partial"`
}

func runStreamingChat(connID string, msg types.WebviewMessage, jsonMode bool) (*StreamResult, error) {
	c, err := client.NewWithOverrides("", "")
	if err != nil {
		return nil, err
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

	var capturedTaskID string
	var sessionStatus string
	var lastAskType string

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

				if payload.TaskID != "" {
					capturedTaskID = payload.TaskID
				}

				m := payload.Message

				if m.Type == "ask" && !m.Partial {
					sessionStatus = "asking"
					lastAskType = m.Ask
					if jsonMode {
						emitJSON(StreamEvent{TaskID: capturedTaskID, Ts: m.Ts, Type: m.Type, Ask: m.Ask, Text: m.Text, Partial: m.Partial})
					}
					return false
				}

				if m.Type != "say" || skipSayTypes[m.Say] {
					return true
				}
				if !printableSayTypes[m.Say] {
					return true
				}

				text := m.Text
				if m.Say == "reasoning" && m.Reasoning != "" {
					text = m.Reasoning
				}

				if jsonMode {
					emitJSON(StreamEvent{TaskID: capturedTaskID, Ts: m.Ts, Type: m.Type, Say: m.Say, Text: m.Text, Reasoning: m.Reasoning, Partial: m.Partial})
				} else {
					lastLen := textLenByTs[m.Ts]
					if text != "" && len(text) > lastLen {
						delta := text[lastLen:]
						fmt.Print(delta)
						textLenByTs[m.Ts] = len(text)
					}
				}

			case "message.add":
				if err := json.Unmarshal([]byte(event.Data), &payload); err != nil {
					return true
				}

				if payload.TaskID != "" {
					capturedTaskID = payload.TaskID
				}

				m := payload.Message

				if m.Type == "say" && printableSayTypes[m.Say] {
					if jsonMode {
						emitJSON(StreamEvent{TaskID: capturedTaskID, Ts: m.Ts, Type: m.Type, Say: m.Say, Text: m.Text, Reasoning: m.Reasoning, Partial: m.Partial})
					} else {
						text := m.Text
						lastLen := textLenByTs[m.Ts]
						if text != "" && len(text) > lastLen {
							fmt.Print(text[lastLen:])
						}
					}
				}

				if m.Type == "ask" && !m.Partial {
					sessionStatus = "asking"
					lastAskType = m.Ask
					if jsonMode {
						emitJSON(StreamEvent{TaskID: capturedTaskID, Ts: m.Ts, Type: m.Type, Ask: m.Ask, Text: m.Text, Partial: m.Partial})
					}
					return false
				}
				if m.Say == "completion_result" {
					sessionStatus = "completed"
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
						if jsonMode {
							emitJSON(map[string]string{"type": "error", "title": notif.Title, "message": notif.Message})
						} else {
							fmt.Fprintf(os.Stderr, "\n[%s] %s: %s\n", notif.Type, notif.Title, notif.Message)
						}
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
		return nil, fmt.Errorf("SSE connection failed before sending message: %w", err)
	case <-time.After(10 * time.Second):
		cancel()
		return nil, fmt.Errorf("SSE connection timeout (10s), server may be unreachable")
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
	case postErr := <-postDone:
		if postErr != nil {
			cancel()
			<-sseDone
			fmt.Println()
			return nil, fmt.Errorf("failed to send message: %w", postErr)
		}
		select {
		case sseErr := <-sseDone:
			if sseErr != nil && sseErr != context.Canceled {
				output.PrintError("SSE stream error: %v", sseErr)
			}
		case <-ctx.Done():
		}
	case <-ctx.Done():
	}

	if !jsonMode {
		fmt.Println()
	}

	if capturedTaskID == "" && msg.TaskID != "" {
		capturedTaskID = msg.TaskID
	}

	return &StreamResult{
		TaskID:      capturedTaskID,
		ConnID:      connID,
		Status:      sessionStatus,
		LastAskType: lastAskType,
	}, nil
}

func emitJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}
