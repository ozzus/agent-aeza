package checks

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type TCPChecker struct {
	baseMetadata
	timeout time.Duration
}

func NewTCPChecker(timeout time.Duration, location, country string) *TCPChecker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &TCPChecker{
		baseMetadata: newBaseMetadata(location, country),
		timeout:      timeout,
	}
}

func (t *TCPChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	address, err := t.resolveAddress(target, parameters)
	if err != nil {
		return &domain.CheckResult{Status: domain.StatusFailed, Error: err.Error()}, nil
	}

	timeout := durationParam(parameters, "timeout", t.timeout)
	if timeout <= 0 {
		timeout = t.timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	d := net.Dialer{}
	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", address)
	duration := time.Since(start)

	hostOnly := address
	if host, _, splitErr := net.SplitHostPort(address); splitErr == nil {
		hostOnly = host
	}

	if err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}

		payload := map[string]interface{}{
			"tcp": []map[string]interface{}{
				{
					"location":    t.locationValue(parameters),
					"country":     t.countryValue(parameters),
					"connectTime": formatSeconds(duration),
					"status":      "Failed",
					"ip":          hostOnly,
				},
			},
		}

		return &domain.CheckResult{
			Status:  domain.StatusFailed,
			Error:   err.Error(),
			Payload: payload,
		}, nil
	}
	defer conn.Close()

	remoteAddr := conn.RemoteAddr()
	ip := hostOnly
	if tcpAddr, ok := remoteAddr.(*net.TCPAddr); ok {
		ip = tcpAddr.IP.String()
	}

	duration = time.Since(start)

	payload := map[string]interface{}{
		"tcp": []map[string]interface{}{
			{
				"location":    t.locationValue(parameters),
				"country":     t.countryValue(parameters),
				"connectTime": formatSeconds(duration),
				"status":      "Connected",
				"ip":          ip,
			},
		},
	}

	return &domain.CheckResult{
		Status:  domain.StatusSuccess,
		Payload: payload,
	}, nil
}

func (t *TCPChecker) resolveAddress(target string, parameters map[string]interface{}) (string, error) {
	if target == "" {
		return "", fmt.Errorf("empty target")
	}

	if strings.Contains(target, "://") {
		parsed, err := url.Parse(target)
		if err != nil {
			return "", err
		}

		host := parsed.Hostname()
		port := parsed.Port()
		if port == "" {
			switch parsed.Scheme {
			case "https":
				port = "443"
			case "http":
				port = "80"
			default:
				port = stringParam(parameters, "port", "80")
			}
		}

		return net.JoinHostPort(host, port), nil
	}

	if strings.Contains(target, ":") {
		return target, nil
	}

	port := stringParam(parameters, "port", "")
	if port == "" {
		port = "80"
	}

	return net.JoinHostPort(target, port), nil
}

func (t *TCPChecker) Type() domain.TaskType {
	return domain.TaskTypeTCP
}
