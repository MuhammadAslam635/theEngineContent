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
		return
	}
	log.Println("Database connection established")

	// AutoMigrate runs in dependency order so foreign keys resolve correctly.
	// Add new models here as they are introduced — never remove existing ones.
	err = DB.AutoMigrate(
		// --- Existing core tables ---
		&models.User{},
		&models.AuditLog{},
		&models.AiTask{},

		// --- Content Engine: reference / seed tables first ---
		&models.Platform{},
		&models.Persona{},
		&models.NickCredential{},

		// --- Intelligence layer ---
		&models.CompetitorAccount{},
		&models.OutlierReel{},
		&models.HookLibrary{},
		&models.AngleLibrary{},
		&models.TrendSignal{},

		// --- Script storehouse ---
		&models.ApprovedScript{},

		// --- Pipeline ---
		&models.ContentBrief{},
		&models.ScriptGenerationAttempt{},
		&models.SSMLScript{},
		&models.AgentConfidenceLog{},

		// --- Production ---
		&models.ContentVideo{},
		&models.VideoScene{},
		&models.PostingPackage{},
		&models.VideoAnalytic{},
	)
	if err != nil {
		log.Printf("Warning: AutoMigrate failed: %v\n", err)
	} else {
		log.Println("AutoMigrate completed successfully")
	}
}
