package main

import (
	"fmt"

	"backend-go/config"
	"backend-go/db"
	httpclient "backend-go/http-client"
	"backend-go/internal/handlers"
	"backend-go/internal/repositories"
	"backend-go/internal/routes"
	"backend-go/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db.InitDB(cfg)

	// Initialize Repositories
	authRepo := repositories.NewAuthRepository(db.DB)

	// Initialize Services
	authService := services.NewAuthService(authRepo, cfg)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize HTTP Client
	client := httpclient.NewHTTPClient(cfg)

	r := gin.Default()

	// Register Routes
	routes.RegisterRoutes(r, cfg, client, authHandler)

	fmt.Printf("Starting Backend Go on port %s...\n", cfg.Port)
	r.Run(":" + cfg.Port)
}
