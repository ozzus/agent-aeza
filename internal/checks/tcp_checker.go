package checks

import (
	"context"
	"fmt"
	"net"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type TCPChecker struct{}

func (c *TCPChecker) Run(ctx context.Context, task domain.Task, agentID string) Result {
	start := time.Now()
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	target := task.Target
	// target expected like "host:port" — если в задачах порт в parameters, используйте его.
	if _, ok := task.Parameters["port"]; ok {
		port := fmt.Sprintf("%v", task.Parameters["port"])
		target = net.JoinHostPort(task.Target, port)
	}

	d := &net.Dialer{}
	connCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := d.DialContext(connCtx, "tcp", target)
	if err != nil {
		return Result{
			TaskID:   task.ID,
			AgentID:  agentID,
			Type:     domain.TaskTypeTCP,
			Target:   target,
			OK:       false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	_ = conn.Close()

	return Result{
		TaskID:   task.ID,
		AgentID:  agentID,
		Type:     domain.TaskTypeTCP,
		Target:   target,
		OK:       true,
		Duration: time.Since(start),
	}
}
