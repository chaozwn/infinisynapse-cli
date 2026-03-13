package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/types"
)

type Client struct {
	baseURL    string
	token      string
	lang       string
	httpClient *http.Client
}

func New() (*Client, error) {
	server := config.GetServer()
	if server == "" {
		return nil, fmt.Errorf("server not configured. Run: isc auth login --server <URL> --token <TOKEN>")
	}

	token := config.GetToken()
	if token == "" {
		return nil, fmt.Errorf("token not configured. Run: isc auth login --server <URL> --token <TOKEN>")
	}

	server = strings.TrimRight(server, "/")

	return &Client{
		baseURL: server,
		token:   token,
		lang:    config.GetLang(),
		httpClient: &http.Client{
			Timeout: 100 * time.Second,
		},
	}, nil
}

func NewWithOverrides(server, token string) (*Client, error) {
	if server == "" {
		server = config.GetServer()
	}
	if token == "" {
		token = config.GetToken()
	}
	if server == "" {
		return nil, fmt.Errorf("server not configured")
	}

	server = strings.TrimRight(server, "/")

	return &Client{
		baseURL: server,
		token:   token,
		lang:    config.GetLang(),
		httpClient: &http.Client{
			Timeout: 100 * time.Second,
		},
	}, nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	u := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		token := c.token
		if !strings.HasPrefix(token, "Bearer ") {
			token = "Bearer " + token
		}
		req.Header.Set("Authorization", token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-lang", c.lang)

	return req, nil
}

func (c *Client) Do(method, path string, body interface{}) (json.RawMessage, error) {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp types.APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return respBody, nil
	}

	if apiResp.Code != 200 {
		if apiResp.Code == 1101 || apiResp.Code == 1105 {
			return nil, fmt.Errorf("token expired or invalid (code: %d). Please re-login: isc auth login", apiResp.Code)
		}
		return nil, fmt.Errorf("API error (code: %d): %s", apiResp.Code, apiResp.Message)
	}

	return apiResp.Data, nil
}

func (c *Client) Get(path string, params map[string]string) (json.RawMessage, error) {
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			if v != "" {
				values.Set(k, v)
			}
		}
		if encoded := values.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
	}
	return c.Do(http.MethodGet, path, nil)
}

func (c *Client) Post(path string, body interface{}) (json.RawMessage, error) {
	return c.Do(http.MethodPost, path, body)
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) Token() string {
	return c.token
}

func (c *Client) RawRequest(method, path string, body interface{}) (*http.Response, error) {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}
