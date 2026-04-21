package config

import "os"

type Config struct {
	Port               string
	DBHost             string
	DBPort             string
	DBName             string
	DBUser             string
	DBPassword         string
	AIOrchestrationURL string
	SociaVaultAPIKey   string
	SociaVaultBaseURL  string
}

func LoadConfig() *Config {
	return &Config{
		Port:               getEnv("APP_PORT", "9001"),
		DBHost:             getEnv("DB_HOST", "postgres-service"),
		DBPort:             getEnv("DB_PORT", "5236"),
		DBName:             getEnv("DB_NAME", "contgen"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		AIOrchestrationURL: getEnv("AI_ORCHESTRATION_URL", "http://ai-orchestration-service:8086"),
		SociaVaultAPIKey:   getEnv("SOCIAVAULT_API", ""),
		SociaVaultBaseURL:  getEnv("SOCIAVAULT_BASE_URL", "https://api.sociavault.com/v1/scrape/"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
