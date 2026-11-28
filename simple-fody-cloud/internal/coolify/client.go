// Package coolify is a tiny client for the parts of Coolify's HTTP API
// that FodyCloud needs. Everything that talks to Coolify goes through here,
// so the rest of the code never builds URLs or sets auth headers itself.
package coolify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client holds the connection details for one Coolify instance.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New creates a client. baseURL is like "http://192.168.122.236:8000".
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// APIError is returned when Coolify answers with a non-2xx status.
// Keeping the status and body lets callers react (e.g. 409 = domain taken).
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("coolify returned %d: %s", e.Status, e.Body)
}

// do sends one request. in (if not nil) is encoded as the JSON body,
// out (if not nil) receives the decoded JSON response.
func (c *Client) do(method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		payload, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, c.baseURL+"/api/v1"+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &APIError{Status: resp.StatusCode, Body: string(data)}
	}

	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("parse response: %w (body: %s)", err, data)
		}
	}
	return nil
}