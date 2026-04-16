package db

import (
	"fmt"
	"log"

	"backend-go/config"
	"backend-go/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v\n", err)
	} else {
		log.Println("Database connection established")

		// Run AutoMigrate
		err = DB.AutoMigrate(&models.User{}, &models.AuditLog{}, &models.AiTask{})
		if err != nil {
			log.Printf("Warning: AutoMigrate failed: %v\n", err)
		} else {
			log.Println("AutoMigrate completed successfully")
		}
	}
}
