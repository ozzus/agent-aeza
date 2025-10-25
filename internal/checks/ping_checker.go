package checks

import (
	"context"
	"time"

	"ozzus/agent-aeza/internal/domain"

	"github.com/go-ping/ping"
)

type PingChecker struct{}

func (c *PingChecker) Run(ctx context.Context, task domain.Task, agentID string) Result {
	start := time.Now()
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	target := task.Target
	pinger, err := ping.NewPinger(target)
	if err != nil {
		return Result{
			TaskID:   task.ID,
			AgentID:  agentID,
			Type:     domain.TaskTypePing,
			Target:   target,
			OK:       false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	pinger.Count = 3
	pinger.Timeout = timeout

	done := make(chan struct{})
	var runErr error
	go func() {
		runErr = pinger.Run()
		close(done)
	}()

	select {
	case <-ctx.Done():
		pinger.Stop()
		<-done
		return Result{
			TaskID:   task.ID,
			AgentID:  agentID,
			Type:     domain.TaskTypePing,
			Target:   target,
			OK:       false,
			Duration: time.Since(start),
			Error:    ctx.Err().Error(),
		}
	case <-done:
		// ok
	}

	if runErr != nil {
		return Result{
			TaskID:   task.ID,
			AgentID:  agentID,
			Type:     domain.TaskTypePing,
			Target:   target,
			OK:       false,
			Duration: time.Since(start),
			Error:    runErr.Error(),
		}
	}

	stats := pinger.Statistics()
	meta := map[string]interface{}{
		"packets_sent":  stats.PacketsSent,
		"packets_recv":  stats.PacketsRecv,
		"avg_rtt_ms":    stats.AvgRtt.Milliseconds(),
		"min_rtt_ms":    stats.MinRtt.Milliseconds(),
		"max_rtt_ms":    stats.MaxRtt.Milliseconds(),
		"stddev_rtt_ms": stats.StdDevRtt.Milliseconds(),
		"ip_addr":       stats.IPAddr,
	}

	return Result{
		TaskID:   task.ID,
		AgentID:  agentID,
		Type:     domain.TaskTypePing,
		Target:   target,
		OK:       stats.PacketsRecv > 0,
		Duration: time.Since(start),
		Meta:     meta,
	}
}
