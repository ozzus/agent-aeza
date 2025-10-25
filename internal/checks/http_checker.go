package checks

import (
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type HTTPChecker struct{}

func (h *HTTPChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	start := time.Now()
	time.Sleep(100 * time.Millisecond)
	duration := time.Since(start)

	statusCode := intParam(parameters, "status", 200)
	resultText := stringParam(parameters, "result", "")
	if resultText == "" {
		if statusCode >= 200 && statusCode < 400 {
			resultText = "OK"
		} else {
			resultText = "FAILED"
		}
	}

	httpPayload := []map[string]interface{}{
		{
			"location": stringParam(parameters, "location", "Unknown"),
			"country":  lowerStringParam(parameters, "country", ""),
			"time":     formatSeconds(duration),
			"status":   statusCode,
			"ip":       stringParam(parameters, "ip", target),
			"result":   resultText,
		},
	}

	return &domain.CheckResult{
		Status:  domain.StatusSuccess,
		Error:   "",
		Payload: map[string]interface{}{"http": httpPayload},
	}, nil
}

func (h *HTTPChecker) Type() domain.TaskType {
	return domain.TaskTypeHTTP
}
