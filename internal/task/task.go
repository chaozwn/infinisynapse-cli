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

type streamState struct {
	capturedTaskID string
	lastAskType    string
	lastMessage    *StreamEvent
}

type streamEventPayload struct {
	TaskID  string      `json:"taskId"`
	Message StreamEvent `json:"message"`
}

func (s *streamState) updateTaskID(taskID string) {
	if taskID != "" {
		s.capturedTaskID = taskID
	}
}

func (s *streamState) rememberMessage(m StreamEvent) {
	m.TaskID = s.capturedTaskID
	s.lastMessage = &m
}

func handleStreamEvent(event client.SSEEvent, state *streamState) bool {
	if event.Event == "heartbeat" {
		return true
	}

	switch event.Event {
	case "message.partial", "message.add":
		var payload streamEventPayload
		if err := json.Unmarshal([]byte(event.Data), &payload); err != nil {
			return true
		}

		state.updateTaskID(payload.TaskID)
		m := payload.Message

		if m.Type == "say" {
			state.rememberMessage(m)
			if event.Event == "message.add" && m.Say == "completion_result" && !m.Partial {
				return false
			}
			return true
		}

		if m.Type == "ask" && !m.Partial {
			if m.Ask != "completion_result" {
				state.rememberMessage(m)
			}
			state.lastAskType = m.Ask
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

	state := &streamState{}

	go func() {
		err := sseClient.Subscribe(ctx, fmt.Sprintf("/api/ai/events?connId=%s", connID), sseReady, func(event client.SSEEvent) bool {
			return handleStreamEvent(event, state)
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

	if state.capturedTaskID == "" && msg.TaskID != "" {
		state.capturedTaskID = msg.TaskID
	}

	return &StreamResult{
		TaskID:      state.capturedTaskID,
		ConnID:      connID,
		LastAskType: state.lastAskType,
		LastMessage: state.lastMessage,
	}, nil
}
