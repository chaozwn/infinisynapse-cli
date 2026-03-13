package client

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
)

type SSEEvent struct {
	Event string
	Data  string
	ID    string
}

type SSEClient struct {
	client *Client
}

func NewSSEClient(c *Client) *SSEClient {
	return &SSEClient{client: c}
}

func (s *SSEClient) Subscribe(ctx context.Context, path string, handler func(event SSEEvent) bool) error {
	u := s.client.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	if s.client.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.client.token)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSE connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE returned HTTP %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var current SSEEvent
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		if line == "" {
			if current.Data != "" || current.Event != "" {
				shouldContinue := handler(current)
				if !shouldContinue {
					return nil
				}
				current = SSEEvent{}
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")

		switch field {
		case "event":
			current.Event = value
		case "data":
			if current.Data != "" {
				current.Data += "\n"
			}
			current.Data += value
		case "id":
			current.ID = value
		}
	}

	return scanner.Err()
}
