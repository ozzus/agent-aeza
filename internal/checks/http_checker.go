package checks

import (
	"context"
	"net/http"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type HTTPChecker struct{}

func (c *HTTPChecker) Run(ctx context.Context, task domain.Task, agentID string) Result {
	start := time.Now()
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := task.Target
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return Result{
			TaskID:   task.ID,
			AgentID:  agentID,
			Type:     domain.TaskTypeHTTP,
			Target:   url,
			OK:       false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{
			TaskID:   task.ID,
			AgentID:  agentID,
			Type:     domain.TaskTypeHTTP,
			Target:   url,
			OK:       false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	defer resp.Body.Close()

	ok := resp.StatusCode >= 200 && resp.StatusCode < 400
	meta := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
	}

	return Result{
		TaskID:   task.ID,
		AgentID:  agentID,
		Type:     domain.TaskTypeHTTP,
		Target:   url,
		OK:       ok,
		Duration: time.Since(start),
		Meta:     meta,
	}
}
