package checks

import (
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type TracerouteChecker struct{}

func (t *TracerouteChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	// Временная заглушка
	time.Sleep(200 * time.Millisecond)

	return &domain.CheckResult{
		Status: domain.StatusSuccess,
	}, nil
}

func (t *TracerouteChecker) Type() domain.TaskType {
	return domain.TaskTypeTraceroute
}
