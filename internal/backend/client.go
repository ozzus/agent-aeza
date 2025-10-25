package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client executes authenticated requests against the backend API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	agentName  string
	agentToken string
}

// NewClient constructs a backend client that authenticates using Basic Auth.
func NewClient(baseURL, agentName, agentToken string) (*Client, error) {
	normalizedURL, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	if agentName == "" {
		return nil, errors.New("agent name is required")
	}

	if agentToken == "" {
		return nil, errors.New("agent token is required")
	}

	return &Client{
		baseURL: normalizedURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		agentName:  agentName,
		agentToken: agentToken,
	}, nil
}

// WithHTTPClient overrides the default http.Client. Primarily useful for testing.
func (c *Client) WithHTTPClient(httpClient *http.Client) {
	if httpClient != nil {
		c.httpClient = httpClient
	}
}

func (c *Client) Heartbeat(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodPost, "/api/agents/heartbeat", nil)
	if err != nil {
		return fmt.Errorf("create heartbeat request: %w", err)
	}

	return c.do(req, nil)
}

func normalizeBaseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("backend base URL is required")
	}

	if !strings.Contains(trimmed, "://") {
		trimmed = "http://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid backend base URL: %w", err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid backend base URL: %s", raw)
	}

	parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return strings.TrimSuffix(parsed.String(), "/"), nil
}

func (c *Client) RotateToken(ctx context.Context, agentID string) (string, error) {
	if agentID == "" {
		agentID = c.agentName
	}

	path := fmt.Sprintf("/api/agents/%s/rotate-token", agentID)
	req, err := c.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return "", fmt.Errorf("create rotate-token request: %w", err)
	}

	var payload struct {
		Token string `json:"token"`
	}

	if err := c.do(req, &payload); err != nil {
		return "", err
	}

	if payload.Token == "" {
		return "", errors.New("rotate token response did not contain token")
	}

	return payload.Token, nil
}

func (c *Client) SendCheckResult(ctx context.Context, checkID string, body io.Reader) error {
	if checkID == "" {
		return errors.New("check ID is required")
	}

	path := fmt.Sprintf("/api/check/%s/results", checkID)
	req, err := c.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return fmt.Errorf("create send result request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return c.do(req, nil)
}

func (c *Client) SendCheckLog(ctx context.Context, checkID string, body io.Reader) error {
	if checkID == "" {
		return errors.New("check ID is required")
	}

	path := fmt.Sprintf("/api/check/%s/logs", checkID)
	req, err := c.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return fmt.Errorf("create send log request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return c.do(req, nil)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.agentName, c.agentToken)
	return req, nil
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			host := req.URL.Hostname()
			return fmt.Errorf("execute request: network error contacting %s: %w", host, err)
		}
		if urlErr, ok := err.(*url.Error); ok {
			return fmt.Errorf("execute request: %s", urlErr.Err)
		}
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		if len(b) == 0 {
			return fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
