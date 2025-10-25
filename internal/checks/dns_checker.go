package checks

import (
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type DNSChecker struct{}

func (d *DNSChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	// Временная заглушка
	time.Sleep(60 * time.Millisecond)

	return &domain.CheckResult{
		Status: domain.StatusSuccess,
	}, nil
}

func (d *DNSChecker) Type() domain.TaskType {
	return domain.TaskTypeDNS
}
