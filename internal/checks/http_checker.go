package checks

import (
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type HTTPChecker struct{}

func (h *HTTPChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	// Временная заглушка - всегда успех
	time.Sleep(100 * time.Millisecond) // Имитация работы

	return &domain.CheckResult{
		Status: domain.StatusSuccess,
		Error:  "",
	}, nil
}

func (h *HTTPChecker) Type() domain.TaskType {
	return domain.TaskTypeHTTP
}
