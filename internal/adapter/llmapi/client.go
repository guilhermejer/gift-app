package llmapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Message    string
	Details    any
}

func (e *APIError) Error() string {
	if e == nil {
		return "llm api error"
	}
	return fmt.Sprintf("llm api error: status=%d message=%s", e.StatusCode, e.Message)
}

type pythonErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Details any    `json:"details"`
	} `json:"error"`
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Chat(ctx context.Context, requestID string, payload any) (map[string]any, error) {
	return c.doJSON(ctx, http.MethodPost, "/profiles/agent/chat", requestID, payload)
}

func (c *Client) Finalize(ctx context.Context, requestID string, payload any) (map[string]any, error) {
	return c.doJSON(ctx, http.MethodPost, "/profiles/agent/finalize", requestID, payload)
}

func (c *Client) DeleteSession(ctx context.Context, requestID, friendID string) (map[string]any, error) {
	path := "/profiles/agent/session/" + friendID
	return c.doJSON(ctx, http.MethodDelete, path, requestID, nil)
}

func (c *Client) Health(ctx context.Context, requestID string) (map[string]any, error) {
	return c.doJSON(ctx, http.MethodGet, "/health", requestID, nil)
}

func (c *Client) SuggestionChat(ctx context.Context, requestID string, payload any) (map[string]any, error) {
	return c.doJSON(ctx, http.MethodPost, "/suggestions/agent/chat", requestID, payload)
}

func (c *Client) SuggestionCreate(ctx context.Context, requestID, friendID string, payload any) (map[string]any, error) {
	path := "/profiles/" + friendID + "/suggestions"
	return c.doJSON(ctx, http.MethodPost, path, requestID, payload)
}

func (c *Client) SuggestionFinalize(ctx context.Context, requestID string, payload any) (map[string]any, error) {
	return c.doJSON(ctx, http.MethodPost, "/suggestions/agent/finalize", requestID, payload)
}

func (c *Client) doJSON(ctx context.Context, method, path, requestID string, body any) (map[string]any, error) {
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if requestID != "" {
		req.Header.Set("X-Request-Id", requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, respBody)
	}

	if len(respBody) == 0 {
		return map[string]any{"status": "ok"}, nil
	}

	var parsed map[string]any
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func parseAPIError(statusCode int, body []byte) error {
	var errResp pythonErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		return &APIError{StatusCode: statusCode, Message: errResp.Error.Message, Details: errResp.Error.Details}
	}

	return &APIError{StatusCode: statusCode, Message: string(body)}
}
