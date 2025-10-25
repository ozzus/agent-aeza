package checks

import (
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type TCPChecker struct{}

func (t *TCPChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	// Временная заглушка
	time.Sleep(80 * time.Millisecond)

	return &domain.CheckResult{
		Status: domain.StatusSuccess,
	}, nil
}

func (t *TCPChecker) Type() domain.TaskType {
	return domain.TaskTypeTCP
}
