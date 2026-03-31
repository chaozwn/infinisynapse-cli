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
	LastAskType string
	LastMessage *StreamEvent
}

func RunNewTask(query string) (*StreamResult, error) {
	connID := uuid.New().String()
	msg := types.WebviewMessage{
		Type:   "newTask",
		Text:   query,
		ConnID: connID,
	}
	return runStreamingChat(connID, msg)
}

func RunAskResponse(taskID, query string) (*StreamResult, error) {
	connID := uuid.New().String()
	msg := types.WebviewMessage{
		Type:        "askResponse",
		Text:        query,
		ConnID:      connID,
		TaskID:      taskID,
		AskResponse: "messageResponse",
	}
	return runStreamingChat(connID, msg)
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

func runStreamingChat(connID string, msg types.WebviewMessage) (*StreamResult, error) {
	c, err := client.NewWithOverrides("", "")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	sseClient := client.NewSSEClient(c)
	sseDone := make(chan error, 1)
	sseReady := make(chan struct{})

	var capturedTaskID string
	var lastAskType string
	var lastMessage *StreamEvent

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
				lastAskType = m.Ask
				return false
			}

			if m.Type == "say" && m.Ask != "completion_result" {
				ev := &StreamEvent{TaskID: capturedTaskID, Ts: m.Ts, Type: m.Type, Say: m.Say, Text: m.Text, Reasoning: m.Reasoning, Partial: m.Partial}
				lastMessage = ev
			}

		case "message.add":
			if err := json.Unmarshal([]byte(event.Data), &payload); err != nil {
				return true
			}

			if payload.TaskID != "" {
				capturedTaskID = payload.TaskID
			}

			m := payload.Message

			if !(m.Type == "ask" && m.Ask == "completion_result") {
				ev := &StreamEvent{TaskID: capturedTaskID, Ts: m.Ts, Type: m.Type, Say: m.Say, Ask: m.Ask, Text: m.Text, Reasoning: m.Reasoning, Partial: m.Partial}
				lastMessage = ev
			}

			if m.Type == "ask" && !m.Partial {
				lastAskType = m.Ask
				return false
			}
			if m.Say == "completion_result" {
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

	if capturedTaskID == "" && msg.TaskID != "" {
		capturedTaskID = msg.TaskID
	}

	return &StreamResult{
		TaskID:      capturedTaskID,
		ConnID:      connID,
		LastAskType: lastAskType,
		LastMessage: lastMessage,
	}, nil
}

