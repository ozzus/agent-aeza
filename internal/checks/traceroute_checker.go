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
	tracerouteFromRegexp  = regexp.MustCompile(`(?m)^From ([^\s]+)`)
	tracerouteBytesRegexp = regexp.MustCompile(`(?m)^\d+ bytes from ([^\s]+)(?: \(([0-9a-fA-F:\.]+)\))?`)
	tracerouteTimeRegexp  = regexp.MustCompile(`time[=<]([0-9.]+) ms`)
)

type TracerouteChecker struct {
	baseMetadata
	maxHops int
	timeout time.Duration
}

func NewTracerouteChecker(maxHops int, timeout time.Duration, location, country string) *TracerouteChecker {
	if maxHops <= 0 {
		maxHops = 30
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	return &TracerouteChecker{
		baseMetadata: newBaseMetadata(location, country),
		maxHops:      maxHops,
		timeout:      timeout,
	}
}

func (t *TracerouteChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	if target == "" {
		return &domain.CheckResult{Status: domain.StatusFailed, Error: "empty target"}, nil
	}

	maxHops := intParam(parameters, "max_hops", t.maxHops)
	if maxHops <= 0 {
		maxHops = t.maxHops
	}

	hopTimeout := durationParam(parameters, "timeout", t.timeout)
	if hopTimeout <= 0 {
		hopTimeout = t.timeout
	}

	var hops []map[string]interface{}
	status := domain.StatusSuccess
	var errText string

	for ttl := 1; ttl <= maxHops; ttl++ {
		hop, reached, hopErr := t.runHop(target, ttl, hopTimeout)
		hop["hop"] = ttl
		hops = append(hops, hop)

		if hopErr != nil {
			status = domain.StatusFailed
			if errText == "" {
				errText = hopErr.Error()
			}
		}

		if reached {
			break
		}
	}

	payload := map[string]interface{}{
		"traceroute": hops,
	}

	return &domain.CheckResult{
		Status:  status,
		Error:   errText,
		Payload: payload,
	}, nil
}

func (t *TracerouteChecker) runHop(target string, ttl int, timeout time.Duration) (map[string]interface{}, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout+time.Second)
	defer cancel()

	deadline := int(timeout.Seconds())
	if deadline <= 0 {
		deadline = 1
	}

	args := []string{"-n", "-c", "1", "-t", strconv.Itoa(ttl), "-W", fmt.Sprintf("%d", deadline), target}
	cmd := exec.CommandContext(ctx, "ping", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	hop := map[string]interface{}{
		"ip":   "*",
		"time": "timeout",
	}

	if matches := tracerouteBytesRegexp.FindStringSubmatch(outputStr); len(matches) >= 2 {
		if len(matches) >= 3 && matches[2] != "" {
			hop["ip"] = matches[2]
		} else {
			hop["ip"] = strings.Trim(matches[1], "<>")
		}
		if timeMatch := tracerouteTimeRegexp.FindStringSubmatch(outputStr); len(timeMatch) == 2 {
			hop["time"] = fmt.Sprintf("%s ms", timeMatch[1])
		}
		return hop, true, nil
	}

	if matches := tracerouteFromRegexp.FindStringSubmatch(outputStr); len(matches) == 2 {
		hop["ip"] = strings.Trim(matches[1], "<>")
	}

	if timeMatch := tracerouteTimeRegexp.FindStringSubmatch(outputStr); len(timeMatch) == 2 {
		hop["time"] = fmt.Sprintf("%s ms", timeMatch[1])
	}

	if ctx.Err() != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return hop, false, ctx.Err()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return hop, false, err
		}
	}

	return hop, false, nil
}

func (t *TracerouteChecker) Type() domain.TaskType {
	return domain.TaskTypeTraceroute
}
