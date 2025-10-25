package checks

import (
	"context"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type TracerouteChecker struct{}

func (c *TracerouteChecker) Run(ctx context.Context, task domain.Task, agentID string) Result {
	start := time.Now()

	return Result{
		TaskID:   task.ID,
		AgentID:  agentID,
		Type:     domain.TaskTypeTraceroute,
		Target:   task.Target,
		OK:       false,
		Duration: time.Since(start),
		Error:    "traceroute not implemented on this agent (see internal/checks/traceroute_checker.go)",
		Meta:     map[string]interface{}{"note": "implement with aeden/traceroute or system call"},
	}
}
