package checks

import (
	"context"
	"net"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type DNSChecker struct{}

func (c *DNSChecker) Run(ctx context.Context, task domain.Task, agentID string) Result {
	start := time.Now()
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	target := task.Target

	type dnsRes struct {
		val interface{}
		err error
	}

	lookupA := func() ([]net.IP, error) {

		ch := make(chan dnsRes, 1)
		go func() {
			ips, err := net.LookupIP(target)
			ch <- dnsRes{val: ips, err: err}
		}()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-ch:
			if r.err != nil {
				return nil, r.err
			}
			return r.val.([]net.IP), nil
		}
	}

	aRecords, aErr := lookupA()
	meta := map[string]interface{}{}
	if aErr == nil {
		meta["a"] = aRecords
	} else {
		meta["a_error"] = aErr.Error()
	}

	chmx := make(chan dnsRes, 1)
	go func() {
		mx, err := net.LookupMX(target)
		chmx <- dnsRes{val: mx, err: err}
	}()
	select {
	case <-ctx.Done():
		return Result{
			TaskID:   task.ID,
			AgentID:  agentID,
			Type:     domain.TaskTypeDNS,
			Target:   target,
			OK:       false,
			Duration: time.Since(start),
			Error:    ctx.Err().Error(),
		}
	case r := <-chmx:
		if r.err == nil {
			meta["mx"] = r.val
		} else {
			meta["mx_error"] = r.err.Error()
		}
	}

	ok := (aErr == nil)

	return Result{
		TaskID:   task.ID,
		AgentID:  agentID,
		Type:     domain.TaskTypeDNS,
		Target:   target,
		OK:       ok,
		Duration: time.Since(start),
		Meta:     meta,
	}
}
