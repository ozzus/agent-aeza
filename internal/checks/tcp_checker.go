package checks

import (
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type TCPChecker struct{}

func (t *TCPChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	start := time.Now()
	time.Sleep(80 * time.Millisecond)
	duration := time.Since(start)

	tcpPayload := []map[string]interface{}{
		{
			"location":    stringParam(parameters, "location", "Unknown"),
			"country":     lowerStringParam(parameters, "country", ""),
			"connectTime": formatSeconds(duration),
			"status":      stringParam(parameters, "status", "Connected"),
			"ip":          stringParam(parameters, "ip", target),
		},
	}

	return &domain.CheckResult{
		Status:  domain.StatusSuccess,
		Payload: map[string]interface{}{"tcp": tcpPayload},
	}, nil
}

func (t *TCPChecker) Type() domain.TaskType {
	return domain.TaskTypeTCP
}
