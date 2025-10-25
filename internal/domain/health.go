package domain

import "time"

type HealthStatus string

const (
	HealthStatusHealthy HealthStatus = "healthy"
)

type HealthResponse struct {
	Status    HealthStatus `json:"status"`
	Timestamp time.Time    `json:"timestamp"`
	AgentID   string       `json:"agent_id"`
}
