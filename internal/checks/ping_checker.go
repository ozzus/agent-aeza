package checks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

var (
	pingSummaryRegexp = regexp.MustCompile(`(?m)(\d+) packets transmitted, (\d+) received, ([0-9.]+)% packet loss`)
	pingIPRegexp      = regexp.MustCompile(`(?m)^PING [^\(]*\(([0-9a-fA-F:\.]+)\)`)
	pingRttRegexp     = regexp.MustCompile(`(?m)rtt [^=]*= ([0-9.]+)/([0-9.]+)/([0-9.]+)/`)
)

type PingChecker struct {
	baseMetadata
	timeout time.Duration
	count   int
}

func NewPingChecker(timeout time.Duration, count int, location, country string) *PingChecker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if count <= 0 {
		count = 4
	}

	return &PingChecker{
		baseMetadata: newBaseMetadata(location, country),
		timeout:      timeout,
		count:        count,
	}
}

func (p *PingChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	host, err := normalizeHostname(target)
	if err != nil {
		return &domain.CheckResult{Status: domain.StatusFailed, Error: err.Error()}, nil
	}

	count := intParam(parameters, "count", p.count)
	if count <= 0 {
		count = p.count
	}

	timeout := durationParam(parameters, "timeout", p.timeout)
	if timeout <= 0 {
		timeout = p.timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout+time.Second)
	defer cancel()

	args := []string{"-n", "-c", strconv.Itoa(count), "-w", fmt.Sprintf("%d", int(timeout.Seconds())), host}
	cmd := exec.CommandContext(ctx, "ping", args...)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if ctx.Err() != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		err = ctx.Err()
	}

	transmitted, received, loss := p.parseSummary(outputStr)
	minRTT, avgRTT, maxRTT := p.parseRTT(outputStr)
	ip := p.parseIP(outputStr, host)
	lossText := fmt.Sprintf("%.0f%%", loss)

	payload := map[string]interface{}{
		"ping": []map[string]interface{}{
			{
				"location": p.locationValue(parameters),
				"country":  p.countryValue(parameters),
				"ip":       ip,
				"packets": map[string]interface{}{
					"transmitted": transmitted,
					"received":    received,
					"loss":        lossText,
				},
				"roundTrip": map[string]interface{}{
					"min": minRTT,
					"avg": avgRTT,
					"max": maxRTT,
				},
			},
		},
	}

	status := domain.StatusSuccess
	var errText string

	if err != nil || received == 0 {
		status = domain.StatusFailed
		if err != nil {
			errText = err.Error()
		} else {
			errText = "no packets received"
		}
	}

	return &domain.CheckResult{
		Status:  status,
		Error:   errText,
		Payload: payload,
	}, nil
}

func (p *PingChecker) parseSummary(output string) (int, int, float64) {
	matches := pingSummaryRegexp.FindStringSubmatch(output)
	if len(matches) != 4 {
		return p.count, 0, 100
	}

	transmitted, _ := strconv.Atoi(matches[1])
	received, _ := strconv.Atoi(matches[2])
	loss, _ := strconv.ParseFloat(matches[3], 64)

	return transmitted, received, loss
}

func (p *PingChecker) parseRTT(output string) (string, string, string) {
	matches := pingRttRegexp.FindStringSubmatch(output)
	if len(matches) != 4 {
		return "0.0 ms", "0.0 ms", "0.0 ms"
	}

	return fmt.Sprintf("%s ms", matches[1]), fmt.Sprintf("%s ms", matches[2]), fmt.Sprintf("%s ms", matches[3])
}

func (p *PingChecker) parseIP(output, fallback string) string {
	matches := pingIPRegexp.FindStringSubmatch(output)
	if len(matches) == 2 {
		return matches[1]
	}

	// Try to extract from "bytes from" lines
	bytesFromRegexp := regexp.MustCompile(`bytes from ([^\s]+)(?: \(([0-9a-fA-F:\.]+)\))?`)
	matches = bytesFromRegexp.FindStringSubmatch(output)
	if len(matches) >= 3 && matches[2] != "" {
		return matches[2]
	}
	if len(matches) >= 2 {
		return strings.Trim(matches[1], "<>")
	}

	return fallback
}

func (p *PingChecker) Type() domain.TaskType {
	return domain.TaskTypePing
}
