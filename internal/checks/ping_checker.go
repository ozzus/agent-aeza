package checks

import (
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type PingChecker struct{}

func (p *PingChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	// Временная заглушка
	time.Sleep(50 * time.Millisecond)

	return &domain.CheckResult{
		Status: domain.StatusSuccess,
	}, nil
}

func (p *PingChecker) Type() domain.TaskType {
	return domain.TaskTypePing
}
