package checks

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type HTTPChecker struct {
	baseMetadata
	timeout time.Duration
	client  *http.Client
}

func NewHTTPChecker(timeout time.Duration, location, country string) *HTTPChecker {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{Timeout: timeout}

	return &HTTPChecker{
		baseMetadata: newBaseMetadata(location, country),
		timeout:      timeout,
		client:       client,
	}
}

func (h *HTTPChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	method := strings.ToUpper(stringParam(parameters, "method", "GET"))
	body := stringParam(parameters, "body", "")

	resolvedURL, err := h.prepareURL(target)
	if err != nil {
		return h.failureResult(target, time.Duration(0), fmt.Errorf("invalid url: %w", err), parameters), nil
	}

	timeout := durationParam(parameters, "timeout", h.timeout)
	if timeout <= 0 {
		timeout = h.timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, resolvedURL, strings.NewReader(body))
	if err != nil {
		return h.failureResult(resolvedURL, time.Duration(0), err, parameters), nil
	}

	if headers, ok := parameters["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	start := time.Now()
	resp, err := h.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return h.failureResult(resolvedURL, duration, err, parameters), nil
	}
	defer resp.Body.Close()

	// Ensure body is fully read to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	status := domain.StatusSuccess
	resultText := "OK"
	if resp.StatusCode >= http.StatusBadRequest {
		status = domain.StatusFailed
		resultText = "FAILED"
	}

	payload := map[string]interface{}{
		"http": []map[string]interface{}{
			{
				"location": h.locationValue(parameters),
				"country":  h.countryValue(parameters),
				"time":     formatSeconds(duration),
				"status":   resp.StatusCode,
				"ip":       h.lookupIP(resolvedURL),
				"result":   resultText,
			},
		},
	}

	return &domain.CheckResult{
		Status:  status,
		Payload: payload,
	}, nil
}

func (h *HTTPChecker) lookupIP(target string) string {
	parsed, err := url.Parse(target)
	if err != nil {
		return target
	}

	host := parsed.Host
	if host == "" {
		host = target
	}

	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}

	addrs, err := net.LookupIP(host)
	if err != nil || len(addrs) == 0 {
		return host
	}

	return addrs[0].String()
}

func (h *HTTPChecker) failureResult(target string, duration time.Duration, err error, parameters map[string]interface{}) *domain.CheckResult {
	payload := map[string]interface{}{
		"http": []map[string]interface{}{
			{
				"location": h.locationValue(parameters),
				"country":  h.countryValue(parameters),
				"time":     formatSeconds(duration),
				"status":   0,
				"ip":       target,
				"result":   "FAILED",
			},
		},
	}

	return &domain.CheckResult{
		Status:  domain.StatusFailed,
		Error:   err.Error(),
		Payload: payload,
	}
}

func (h *HTTPChecker) prepareURL(target string) (string, error) {
	if target == "" {
		return "", fmt.Errorf("empty target")
	}

	if !strings.Contains(target, "://") {
		target = "http://" + target
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}

	if parsed.Scheme == "" {
		parsed.Scheme = "http"
	}

	return parsed.String(), nil
}

func (h *HTTPChecker) Type() domain.TaskType {
	return domain.TaskTypeHTTP
}
