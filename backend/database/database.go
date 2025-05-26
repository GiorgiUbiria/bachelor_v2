package database

import (
	"fmt"
	"log"
	"os"

	"bachelor_backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect initializes the database connection
func Connect() {
	var err error

	// Get database configuration from environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres123")
	dbname := getEnv("DB_NAME", "bachelor_db")

	// Create connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, password, dbname)

	// Configure GORM logger
	gormLogger := logger.Default
	if getEnv("GO_ENV", "development") == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Connect to database
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	// Auto-migrate all models
	err = AutoMigrate()
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed")
}

// AutoMigrate runs auto-migration for all models
func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.ShoppingCart{},
		&models.CartItem{},
		&models.UserInteraction{},
		&models.SearchQuery{},
		&models.Recommendation{},
		&models.RecommendationFeedback{},
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.ChatbotIntent{},
		&models.UserSession{},
		&models.ProductView{},
		&models.SearchAnalytics{},
		&models.MLModelPerformance{},
	)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
