package task

import (
	"testing"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
)

func TestHandleStreamEventIgnoresPartialCompletionResult(t *testing.T) {
	state := &streamState{}

	shouldContinue := handleStreamEvent(client.SSEEvent{
		Event: "message.add",
		Data:  `{"taskId":"task-1","message":{"ts":1,"type":"say","say":"completion_result","text":"","partial":true}}`,
	}, state)
	if !shouldContinue {
		t.Fatal("partial completion_result should not stop the stream")
	}
	if state.lastMessage == nil || state.lastMessage.Text != "" || !state.lastMessage.Partial {
		t.Fatalf("unexpected partial last message: %#v", state.lastMessage)
	}

	shouldContinue = handleStreamEvent(client.SSEEvent{
		Event: "message.partial",
		Data:  `{"taskId":"task-1","message":{"ts":1,"type":"say","say":"completion_result","text":"done","partial":false}}`,
	}, state)
	if !shouldContinue {
		t.Fatal("final partial update should wait for the terminal ask")
	}
	if state.lastMessage == nil || state.lastMessage.Text != "done" || state.lastMessage.Partial {
		t.Fatalf("unexpected final last message: %#v", state.lastMessage)
	}

	shouldContinue = handleStreamEvent(client.SSEEvent{
		Event: "message.add",
		Data:  `{"taskId":"task-1","message":{"ts":2,"type":"ask","ask":"completion_result","text":"","partial":false}}`,
	}, state)
	if shouldContinue {
		t.Fatal("terminal completion_result ask should stop the stream")
	}
	if state.lastAskType != "completion_result" {
		t.Fatalf("unexpected last ask type: %q", state.lastAskType)
	}
	if state.lastMessage == nil || state.lastMessage.Text != "done" {
		t.Fatalf("terminal ask should preserve final completion text, got %#v", state.lastMessage)
	}
}

func TestHandleStreamEventStopsOnFinalCompletionResultAdd(t *testing.T) {
	state := &streamState{}

	shouldContinue := handleStreamEvent(client.SSEEvent{
		Event: "message.add",
		Data:  `{"taskId":"task-1","message":{"ts":1,"type":"say","say":"completion_result","text":"done","partial":false}}`,
	}, state)
	if shouldContinue {
		t.Fatal("final completion_result add should stop the stream")
	}
	if state.lastMessage == nil || state.lastMessage.Text != "done" || state.lastMessage.TaskID != "task-1" {
		t.Fatalf("unexpected last message: %#v", state.lastMessage)
	}
}
