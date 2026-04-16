package routes

import (
	"net/http"

	"backend-go/config"
	"backend-go/db"
	"backend-go/internal/handlers"
	httpclient "backend-go/http-client"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, client *httpclient.HTTPClient, authHandler *handlers.AuthHandler) {
	// Public Health routes
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.GET("/ready", func(c *gin.Context) {
		if db.DB != nil {
			sqlDB, err := db.DB.DB()
			if err == nil && sqlDB.Ping() == nil {
				c.String(http.StatusOK, "READY")
				return
			}
		}
		c.String(http.StatusServiceUnavailable, "NOT READY")
	})

	// AI Orchestration health
	r.GET("/ai-health", func(c *gin.Context) {
		status, err := client.CheckAIServiceHealth()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "AI Orchestration unreachable",
				"error":  err.Error(),
			})
			return
		}

		if status != http.StatusOK {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "AI Orchestration unhealthy",
				"code":   status,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "AI Orchestration is healthy"})
	})

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/forget-password", authHandler.ForgetPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Backend Go Service (Gin/GORM) is running on port %s", cfg.Port)
	})
}
