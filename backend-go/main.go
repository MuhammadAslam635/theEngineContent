package main

import (
	"fmt"

	"backend-go/config"
	"backend-go/db"
	_ "backend-go/docs"
	httpclient "backend-go/http-client"
	"backend-go/internal/handlers"
	"backend-go/internal/repositories"
	"backend-go/internal/routes"
	"backend-go/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// @title cont-gen Backend API
// @version 1.0
// @description API for the cont-gen Content Engine platform
// @host localhost:9002
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db.InitDB(cfg)

	// Initialize Repositories
	authRepo := repositories.NewAuthRepository(db.DB)
	agentSettingRepo := repositories.NewAgentSettingRepository(db.DB)
	ytChannelRepo := repositories.NewYouTubeChannelRepository(db.DB)

	// Initialize Services
	authService := services.NewAuthService(authRepo, cfg)
	agentSettingService := services.NewAgentSettingService(agentSettingRepo)
	socialProfileService := services.NewSocialProfileService(ytChannelRepo, cfg)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService)
	agentSettingHandler := handlers.NewAgentSettingHandler(agentSettingService)
	socialProfileHandler := handlers.NewSocialProfileHandler(socialProfileService)

	// Initialize HTTP Client
	client := httpclient.NewHTTPClient(cfg)

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "x-user-id"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Register Routes
	routes.RegisterRoutes(r, cfg, client, authHandler, agentSettingHandler, socialProfileHandler)

	fmt.Printf("Starting Backend Go on port %s...\n", cfg.Port)
	r.Run(":" + cfg.Port)
}
