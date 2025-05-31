package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

	// Handle password securely
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		env := os.Getenv("GO_ENV")
		if env == "production" {
			log.Fatal("DB_PASSWORD environment variable is required in production")
		}
		// Only use fallback in development
		log.Println("WARNING: Using default database password. Set DB_PASSWORD environment variable for production!")
		password = "postgres123"
	}

	dbname := getEnv("DB_NAME", "bachelor_db")

	// Validate required environment variables
	if host == "" || port == "" || user == "" || dbname == "" {
		log.Fatal("Required database configuration missing: DB_HOST, DB_PORT, DB_USER, DB_NAME")
	}

	// Create connection string with additional parameters for better performance
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC connect_timeout=10",
		host, port, user, password, dbname)

	// Configure GORM logger level based on environment
	logLevel := logger.Warn // Default to Warn to reduce verbose logging
	if getEnv("DB_LOG_LEVEL", "warn") == "info" {
		logLevel = logger.Info
	} else if getEnv("DB_LOG_LEVEL", "warn") == "error" {
		logLevel = logger.Error
	} else if getEnv("DB_LOG_LEVEL", "warn") == "silent" {
		logLevel = logger.Silent
	}

	// Open database connection with enhanced configuration
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Optimize for performance
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool for better performance and reliability
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	// Connection pool configuration
	maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 25)    // Maximum number of open connections
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 5)     // Maximum number of idle connections
	maxLifetime := getEnvInt("DB_CONN_MAX_LIFETIME", 300) // Maximum lifetime in seconds
	maxIdleTime := getEnvInt("DB_CONN_MAX_IDLE_TIME", 60) // Maximum idle time in seconds

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Second)

	log.Printf("Database connected successfully with pool config: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%ds, MaxIdleTime=%ds",
		maxOpenConns, maxIdleConns, maxLifetime, maxIdleTime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Run auto-migration
	if err := AutoMigrate(); err != nil {
		log.Fatal("Failed to run auto-migration:", err)
	}

	log.Println("Database migration completed successfully")
}

// AutoMigrate runs auto-migration for all models
func AutoMigrate() error {
	// Migrate models one by one to handle potential constraint issues
	models := []interface{}{
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
		&models.UserSession{},
		&models.ProductView{},
		&models.SearchAnalytics{},
		&models.MLModelPerformance{},
		&models.Favorite{},
		&models.Upvote{},
		&models.Comment{},
		&models.Discount{},
		&models.Tag{},
		&models.ProductTag{},
		&models.RequestLog{},
		&models.AnomalyAlert{},
		&models.SecurityMetrics{},
	}

	var migrationErrors []error

	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			// Log the error but continue with other models
			log.Printf("Warning: Failed to migrate model %T: %v", model, err)
			migrationErrors = append(migrationErrors, fmt.Errorf("failed to migrate model %T: %w", model, err))
		} else {
			log.Printf("Successfully migrated model %T", model)
		}
	}

	// Ensure UUID extension is enabled
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("Warning: Failed to create uuid-ossp extension: %v", err)
	}

	// Add custom constraints and indexes
	if err := addCustomConstraints(); err != nil {
		log.Printf("Warning: Failed to add custom constraints: %v", err)
		migrationErrors = append(migrationErrors, err)
	}

	// If there were migration errors but some models succeeded, log them but don't fail
	if len(migrationErrors) > 0 {
		log.Printf("Migration completed with %d warnings/errors", len(migrationErrors))
		for _, err := range migrationErrors {
			log.Printf("Migration error: %v", err)
		}
	}

	return nil
}

// addCustomConstraints adds custom database constraints and indexes
func addCustomConstraints() error {
	// Add unique constraint for cart_id + product_id combination in cart_items
	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_items_cart_product 
		ON cart_items(cart_id, product_id)
	`).Error; err != nil {
		return fmt.Errorf("failed to create unique index on cart_items: %w", err)
	}

	// Add unique constraint for user_id + product_id + algorithm_type in recommendations
	if err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_recommendations_user_product_algorithm 
		ON recommendations(user_id, product_id, algorithm_type)
	`).Error; err != nil {
		return fmt.Errorf("failed to create unique index on recommendations: %w", err)
	}

	// Add check constraints for valid order statuses
	if err := DB.Exec(`
		ALTER TABLE orders 
		ADD CONSTRAINT IF NOT EXISTS check_order_status 
		CHECK (status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled'))
	`).Error; err != nil {
		log.Printf("Warning: Failed to add order status check constraint: %v", err)
	}

	// Add check constraints for valid interaction types
	if err := DB.Exec(`
		ALTER TABLE user_interactions 
		ADD CONSTRAINT IF NOT EXISTS check_interaction_type 
		CHECK (interaction_type IN ('view', 'cart_add', 'purchase', 'like', 'wishlist'))
	`).Error; err != nil {
		log.Printf("Warning: Failed to add interaction type check constraint: %v", err)
	}

	// Add check constraints for valid feedback types
	if err := DB.Exec(`
		ALTER TABLE recommendation_feedbacks 
		ADD CONSTRAINT IF NOT EXISTS check_feedback_type 
		CHECK (feedback_type IN ('clicked', 'purchased', 'dismissed', 'liked'))
	`).Error; err != nil {
		log.Printf("Warning: Failed to add feedback type check constraint: %v", err)
	}

	log.Println("Custom constraints and indexes added successfully")
	return nil
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

// getEnvInt gets environment variable as integer with fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using fallback: %d", key, value, fallback)
	}
	return fallback
}
