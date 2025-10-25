package middleware

// import (
// 	"time"

// 	"log"

// 	"github.com/gin-gonic/gin"
// )

// // Logger middleware для логирования запросов
// func Logger() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		start := time.Now()

// 		// Продолжаем обработку запроса
// 		c.Next()

// 		// Логируем после обработки
// 		duration := time.Since(start)

// 		log.Printf("HTTP %s %s %d %s",
// 			c.Request.Method,
// 			c.Request.URL.Path,
// 			c.Writer.Status(),
// 			duration,
// 		)
// 	}
// }

// // Recovery middleware для обработки паник
// func Recovery() gin.HandlerFunc {
// 	return gin.Recovery()
// }

// // CORS middleware для разрешения кросс-доменных запросов
// func CORS() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
// 		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
// 		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(204)
// 			return
// 		}

// 		c.Next()
// 	}
// }
