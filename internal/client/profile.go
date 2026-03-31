package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type userProfileResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		UserID string `json:"userId"`
	} `json:"data"`
}

// FetchUserID calls the console profile API with Bearer token and returns the userId.
func FetchUserID(consoleURL, apiKey string) (string, error) {
	url := strings.TrimRight(consoleURL, "/") + "/user/profile"
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create profile request: %w", err)
	}

	req.Header.Set("Authorization", BearerToken(apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("profile request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read profile response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("profile API HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result userProfileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse profile response: %w", err)
	}

	if result.Code != 200 {
		return "", fmt.Errorf("profile API error (code: %d): %s", result.Code, result.Message)
	}

	if result.Data.UserID == "" {
		return "", fmt.Errorf("userId not found in profile response")
	}

	return result.Data.UserID, nil
}
