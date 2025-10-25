package http

import (
	"net/http"
	"time"

	"ozzus/agent-aeza/internal/domain"
	"ozzus/agent-aeza/internal/service"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	agentService *service.AgentService
	agentID      string
}

func NewHealthController(agentService *service.AgentService, agentID string) *HealthController {
	return &HealthController{
		agentService: agentService,
		agentID:      agentID,
	}
}

// Health handler для проверки работоспособности агента
func (h *HealthController) Health(c *gin.Context) {

	if err := h.agentService.HealthCheck(c.Request.Context()); err != nil {
		response := domain.HealthResponse{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			AgentID:   h.agentID,
			Message:   err.Error(),
		}
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response := domain.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		AgentID:   h.agentID,
		Message:   "Agent is running",
	}
	c.JSON(http.StatusOK, response)
}

// Status handler для получения детального статуса агента
func (h *HealthController) Status(c *gin.Context) {
	status := h.agentService.GetStatus()
	c.JSON(http.StatusOK, status)
}

// Ready handler для проверки готовности агента к работе
func (h *HealthController) Ready(c *gin.Context) {
	// Проверяем, что сервис запущен и основные компоненты работают
	if err := h.agentService.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"agent":     h.agentID,
			"message":   err.Error(),
			"timestamp": time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"agent":     h.agentID,
		"message":   "Agent is ready to process tasks",
		"timestamp": time.Now(),
	})
}

// Info handler для получения общей информации об агенте
func (h *HealthController) Info(c *gin.Context) {
	status := h.agentService.GetStatus()

	info := gin.H{
		"agent_id":  h.agentID,
		"status":    status,
		"version":   "1.0.0",
		"timestamp": time.Now(),
		"components": []string{
			"task_processor",
			"health_checker",
			"kafka_consumer",
			"result_sender",
		},
	}

	c.JSON(http.StatusOK, info)
}
