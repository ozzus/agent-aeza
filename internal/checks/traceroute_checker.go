package checks

import (
	"fmt"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type TracerouteChecker struct{}

func (t *TracerouteChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	start := time.Now()
	time.Sleep(200 * time.Millisecond)
	duration := time.Since(start)

	hopsCount := intParam(parameters, "hops_count", 3)
	if hopsCount <= 0 {
		hopsCount = 3
	}

	traceroutePayload := make([]map[string]interface{}, 0, hopsCount)
	for i := 1; i <= hopsCount; i++ {
		hopKey := fmt.Sprintf("hop_%d_ip", i)
		traceroutePayload = append(traceroutePayload, map[string]interface{}{
			"hop":  i,
			"ip":   stringParam(parameters, hopKey, target),
			"time": formatMilliseconds(duration + time.Duration(i)*5*time.Millisecond),
		})
	}

	return &domain.CheckResult{
		Status:  domain.StatusSuccess,
		Payload: map[string]interface{}{"traceroute": traceroutePayload},
	}, nil
}

func (t *TracerouteChecker) Type() domain.TaskType {
	return domain.TaskTypeTraceroute
}
