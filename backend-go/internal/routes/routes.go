package routes

import (
	"net/http"

	"backend-go/config"
	"backend-go/db"
	"backend-go/internal/handlers"
	httpclient "backend-go/http-client"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, client *httpclient.HTTPClient, authHandler *handlers.AuthHandler, agentSettingHandler *handlers.AgentSettingHandler, socialProfileHandler *handlers.SocialProfileHandler) {
	// Public Health routes
	r.GET("/health", healthCheck)
	r.GET("/ready", readyCheck)
	r.GET("/ai-health", aiHealthCheck(client))

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/forget-password", authHandler.ForgetPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	// Agent Settings routes
	agents := r.Group("/agents")
	{
		agents.GET("/all", agentSettingHandler.GetAll)
		agents.GET("/by-name/:name", agentSettingHandler.GetByName)
		agents.GET("/:id", agentSettingHandler.GetByID)
		agents.POST("/create", agentSettingHandler.Create)
		agents.PUT("/update/:id", agentSettingHandler.Update)
		agents.DELETE("/delete/:id", agentSettingHandler.Delete)
	}

	// Social Profile routes
	social := r.Group("/social")
	{
		social.POST("/fetch-profile", socialProfileHandler.FetchProfile)
		social.GET("/youtube/channels", socialProfileHandler.GetAllYouTubeChannelsSSE)
	}

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Backend Go Service (Gin/GORM) is running on port %s", cfg.Port)
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// healthCheck returns OK if the service is alive.
// @Summary Health Check
// @Tags Health
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// readyCheck returns READY if the database is reachable.
// @Summary Readiness Check
// @Tags Health
// @Produce plain
// @Success 200 {string} string "READY"
// @Failure 503 {string} string "NOT READY"
// @Router /ready [get]
func readyCheck(c *gin.Context) {
	if db.DB != nil {
		sqlDB, err := db.DB.DB()
		if err == nil && sqlDB.Ping() == nil {
			c.String(http.StatusOK, "READY")
			return
		}
	}
	c.String(http.StatusServiceUnavailable, "NOT READY")
}

// aiHealthCheck checks if the AI Orchestration service is healthy.
// @Summary AI Orchestration Health
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /ai-health [get]
func aiHealthCheck(client *httpclient.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
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
	}
}
