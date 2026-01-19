package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://blazingly.fast"
	submissionPath = "/api/project"
)

var ErrAlreadyRegistered = errors.New("project already submitted")

// Client wraps interactions with the blazingly.fast API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Submission mirrors the backend submission form.
type Submission struct {
	RepoURL         string
	IsBlazinglyFast bool
	Blurb           string
	Hidden          bool
}

// SubmissionResponse is a subset of the API payload.
type SubmissionResponse struct {
	ID      string          `json:"id"`
	Project json.RawMessage `json:"project"`
}

// Error represents a non-409 API error.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return fmt.Sprintf("api request failed with status %d", e.Status)
	}
	return fmt.Sprintf("api request failed: %s (status %d)", e.Message, e.Status)
}

// NewClient builds a client using the provided base URL.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

// Submit sends the submission payload to the API.
func (c *Client) Submit(ctx context.Context, payload Submission) (*SubmissionResponse, error) {
	body := map[string]interface{}{
		"repoUrl":         payload.RepoURL,
		"isBlazinglyFast": payload.IsBlazinglyFast,
		"blurb":           payload.Blurb,
		"hidden":          payload.Hidden,
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+submissionPath, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "bfast-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusConflict {
		return nil, ErrAlreadyRegistered
	}

	if resp.StatusCode >= 400 {
		return nil, &Error{Status: resp.StatusCode, Message: extractMessage(data)}
	}

	var submission SubmissionResponse
	if len(data) > 0 {
		if err := json.Unmarshal(data, &submission); err != nil {
			return nil, err
		}
	}

	return &submission, nil
}

func extractMessage(body []byte) string {
	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &payload); err == nil {
		if payload.Message != "" {
			return payload.Message
		}
		if payload.Error != "" {
			return payload.Error
		}
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return ""
	}

	if len(trimmed) > 256 {
		return trimmed[:256]
	}

	return trimmed
}
