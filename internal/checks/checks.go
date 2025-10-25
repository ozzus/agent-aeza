package checks

import (
	"context"
	"ozzus/agent-aeza/internal/domain"
	"time"
)

type Result struct {
	TaskID   string
	AgentID  string
	Type     domain.TaskType
	Target   string
	OK       bool
	Duration time.Duration
	Error    string
	Meta     map[string]interface{}
}

type Checker interface {
	Run(ctx context.Context, task domain.Task, agentID string) Result
}

func NewChecker(t domain.TaskType) Checker {
	switch t {
	case domain.TaskTypeHTTP:
		return &HTTPChecker{}
	case domain.TaskTypeTCP:
		return &TCPChecker{}
	case domain.TaskTypePing:
		return &PingChecker{}
	case domain.TaskTypeDNS:
		return &DNSChecker{}
	case domain.TaskTypeTraceroute:
		return &TracerouteChecker{}
	}
}
