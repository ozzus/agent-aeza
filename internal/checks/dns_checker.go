package checks

import (
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type DNSChecker struct{}

func (d *DNSChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	time.Sleep(60 * time.Millisecond)

	ttl := stringParam(parameters, "ttl", "")
	if ttl == "" {
		ttl = formatTTL(durationParam(parameters, "ttl_duration", 5*time.Minute))
	}

	dnsPayload := map[string]interface{}{
		"locations": []map[string]interface{}{
			{
				"location": stringParam(parameters, "location", "Unknown"),
				"country":  lowerStringParam(parameters, "country", ""),
				"records":  stringParam(parameters, "records", target),
				"ttl":      ttl,
			},
		},
	}

	return &domain.CheckResult{
		Status:  domain.StatusSuccess,
		Payload: map[string]interface{}{"dns": dnsPayload},
	}, nil
}

func (d *DNSChecker) Type() domain.TaskType {
	return domain.TaskTypeDNS
}
