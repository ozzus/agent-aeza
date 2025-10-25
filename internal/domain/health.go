package domain

import "time"

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

type HealthResponse struct {
	Status    HealthStatus `json:"status"`
	Timestamp time.Time    `json:"timestamp"`
	AgentID   string       `json:"agent_id"`
	Message   string       `json:"message,omitempty"`
}

// HealthCheckRequest запрос для проверки здоровья
type HealthCheckRequest struct {
	// Можно добавить поля для детальной проверки компонентов
	CheckComponents bool `json:"check_components"`
}

// ComponentHealth статус здоровья компонента
type ComponentHealth struct {
	Name    string      `json:"name"`
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// DetailedHealthResponse детальный ответ о здоровье
type DetailedHealthResponse struct {
	Status     HealthStatus      `json:"status"`
	Timestamp  time.Time         `json:"timestamp"`
	AgentID    string            `json:"agent_id"`
	Components []ComponentHealth `json:"components"`
	Version    string            `json:"version"`
}
