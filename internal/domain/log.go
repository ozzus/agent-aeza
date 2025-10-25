package domain

import "time"

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelError LogLevel = "error"
)

type LogEntry struct {
	TaskID    string    `json:"task_id"`
	AgentID   string    `json:"agent_id"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
