package domain

import "time"

type CheckStatus string

const (
	StatusSuccess CheckStatus = "success"
	StatusFailed  CheckStatus = "failed"
	StatusTimeout CheckStatus = "timeout"
)

type CheckResult struct {
	TaskID    string      `json:"task_id"`
	AgentID   string      `json:"agent_id"`
	Status    CheckStatus `json:"status"`
	Duration  int64       `json:"duration"`
	Error     string      `json:"error"`
	Timestamp time.Time   `json:"timestamp"`
}
