package checks

import (
	"fmt"
	"math"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type PingChecker struct{}

func (p *PingChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	start := time.Now()
	time.Sleep(50 * time.Millisecond)
	duration := time.Since(start)

	transmitted := intParam(parameters, "packets_transmitted", 4)
	received := intParam(parameters, "packets_received", transmitted)

	lossPercent := 0.0
	if transmitted > 0 {
		lossPercent = math.Max(0, float64(transmitted-received)/float64(transmitted)*100)
	}

	minRTT := durationParam(parameters, "rtt_min", duration)
	avgRTT := durationParam(parameters, "rtt_avg", duration+5*time.Millisecond)
	maxRTT := durationParam(parameters, "rtt_max", duration+10*time.Millisecond)

	pingPayload := []map[string]interface{}{
		{
			"location": stringParam(parameters, "location", "Unknown"),
			"country":  lowerStringParam(parameters, "country", ""),
			"ip":       stringParam(parameters, "ip", target),
			"packets": map[string]interface{}{
				"transmitted": transmitted,
				"received":    received,
				"loss":        fmt.Sprintf("%.0f%%", lossPercent),
			},
			"roundTrip": map[string]interface{}{
				"min": formatMilliseconds(minRTT),
				"avg": formatMilliseconds(avgRTT),
				"max": formatMilliseconds(maxRTT),
			},
		},
	}

	return &domain.CheckResult{
		Status:  domain.StatusSuccess,
		Payload: map[string]interface{}{"ping": pingPayload},
	}, nil
}

func (p *PingChecker) Type() domain.TaskType {
	return domain.TaskTypePing
}
