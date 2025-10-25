package http

import "github.com/gin-gonic/gin"

func NewRouter(healthController *HealthController) *gin.Engine {
	router := gin.Default()

	router.GET("/health", healthController.Health)
	router.GET("/status", healthController.Status)
	router.GET("/ready", healthController.Ready)
	router.GET("/info", healthController.Info)

	return router
}
