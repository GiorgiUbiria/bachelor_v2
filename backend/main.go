package main

import (
	"log"
	"os"
	"strings"
	"time"

	"bachelor_backend/database"
	_ "bachelor_backend/docs"
	"bachelor_backend/handlers"
	"bachelor_backend/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/gofiber/swagger"
)

// @title Bachelor E-commerce API
// @version 1.0
// @description ML-Powered E-commerce Platform API with advanced recommendation system, intelligent search, and comprehensive analytics
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@bachelor-ecommerce.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Authentication
// @tag.description User authentication and profile management

// @tag.name Products
// @tag.description Product catalog management and search

// @tag.name Cart
// @tag.description Shopping cart operations

// @tag.name Orders
// @tag.description Order management and tracking

// @tag.name Analytics
// @tag.description Business analytics and reporting

// @tag.name ML
// @tag.description Machine Learning services and model management
func main() {
	// Initialize database connection
	database.Connect()

	// Create Fiber app with enhanced configuration
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
		// Security configurations
		BodyLimit:    4 * 1024 * 1024, // 4MB body limit
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		// Disable server header for security
		DisableStartupMessage: false,
		ServerHeader:          "",
		AppName:               "Bachelor Backend API v1.0",
	})

	// Swagger documentation endpoint
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:          "/swagger/doc.json",
		DeepLinking:  true,
		DocExpansion: "list",
		OAuth: &swagger.OAuthConfig{
			AppName:  "Bachelor E-commerce API",
			ClientId: "bachelor-api-client",
		},
		OAuth2RedirectUrl: "http://localhost:8081/swagger/oauth2-redirect.html",
	}))

	// Security middleware
	app.Use(helmet.New(helmet.Config{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
		HSTSPreloadEnabled:    true,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	// Recovery middleware
	app.Use(recover.New())

	// Structured logging middleware
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${ip} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))

	// Rate limiting middleware
	app.Use(limiter.New(limiter.Config{
		Max:        100,             // 100 requests
		Expiration: 1 * time.Minute, // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Rate limit by IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Rate limit exceeded. Please try again later.",
			})
		},
	}))

	// CORS middleware with enhanced security and Docker support
	app.Use(cors.New(cors.Config{
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Session-ID,X-Requested-With",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
		// Dynamic origin validation for Docker environments
		AllowOriginsFunc: func(origin string) bool {
			// Log the origin for debugging
			log.Printf("CORS: Checking origin: %s", origin)

			// Allow localhost variations
			allowedPatterns := []string{
				"http://localhost:",
				"http://127.0.0.1:",
				"https://localhost:",
				"https://127.0.0.1:",
			}

			for _, pattern := range allowedPatterns {
				if len(origin) > len(pattern) && origin[:len(pattern)] == pattern {
					log.Printf("CORS: Allowed origin (pattern match): %s", origin)
					return true
				}
			}

			// Check environment variable origins
			allowedOrigins := getEnv("ALLOWED_ORIGINS", "")
			if allowedOrigins != "" {
				origins := strings.Split(allowedOrigins, ",")
				for _, allowedOrigin := range origins {
					if strings.TrimSpace(allowedOrigin) == origin {
						log.Printf("CORS: Allowed origin (env match): %s", origin)
						return true
					}
				}
			}

			log.Printf("CORS: Rejected origin: %s", origin)
			return false
		},
	}))

	// Request timeout middleware for long-running requests
	app.Use("/api", timeout.New(func(c *fiber.Ctx) error {
		return c.Next()
	}, 30*time.Second))

	// Health check endpoint with enhanced information
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "bachelor_backend",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
		})
	})

	// CORS test endpoint for debugging
	app.All("/cors-test", func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		method := c.Method()
		log.Printf("CORS Test: Method=%s, Origin=%s", method, origin)

		return c.JSON(fiber.Map{
			"message": "CORS test successful",
			"origin":  origin,
			"method":  method,
			"headers": c.GetReqHeaders(),
		})
	})

	// API routes with versioning
	api := app.Group("/api/v1")

	// Authentication routes with specific rate limiting
	auth := api.Group("/auth")
	auth.Use(limiter.New(limiter.Config{
		Max:        10,              // 10 auth requests
		Expiration: 1 * time.Minute, // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "Too many authentication attempts. Please try again later.",
			})
		},
	}))

	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Get("/profile", middleware.AuthRequired(), handlers.GetProfile)
	auth.Put("/profile", middleware.AuthRequired(), handlers.UpdateProfile)

	// Product routes
	products := api.Group("/products")
	products.Get("/", middleware.OptionalAuth(), handlers.GetProducts)
	products.Get("/categories", handlers.GetCategories)
	products.Get("/search", middleware.OptionalAuth(), handlers.SearchProducts)
	products.Get("/recommendations", middleware.AuthRequired(), handlers.GetRecommendations)
	products.Get("/category/:category", middleware.OptionalAuth(), handlers.GetProductsByCategory)
	products.Get("/:id", middleware.OptionalAuth(), handlers.GetProduct)

	// Admin product management routes
	products.Post("/", middleware.AuthRequired(), handlers.CreateProduct)
	products.Put("/:id", middleware.AuthRequired(), handlers.UpdateProduct)
	products.Delete("/:id", middleware.AuthRequired(), handlers.DeleteProduct)

	// Shopping cart routes
	cart := api.Group("/cart", middleware.AuthRequired())
	cart.Get("/", handlers.GetCart)
	cart.Post("/add", handlers.AddToCart)
	cart.Put("/item/:id", handlers.UpdateCartItem)
	cart.Delete("/item/:id", handlers.RemoveFromCart)
	cart.Delete("/clear", handlers.ClearCart)

	// Order routes
	orders := api.Group("/orders", middleware.AuthRequired())
	orders.Get("/", handlers.GetOrders)
	orders.Get("/stats", handlers.GetOrderStats)
	orders.Get("/:id", handlers.GetOrder)
	orders.Post("/", handlers.CreateOrder)
	orders.Put("/:id/status", handlers.UpdateOrderStatus)
	orders.Put("/:id/cancel", handlers.CancelOrder)

	// Analytics routes
	analytics := api.Group("/analytics", middleware.AuthRequired())
	analytics.Get("/dashboard", handlers.GetDashboard)
	analytics.Get("/user", handlers.GetUserAnalytics)
	analytics.Get("/products", handlers.GetProductAnalytics)
	analytics.Get("/trends", handlers.GetMLTrends)
	analytics.Get("/search", handlers.GetSearchAnalytics)
	analytics.Get("/recommendations/metrics", handlers.GetRecommendationMetrics)
	analytics.Get("/export", handlers.ExportAnalytics)

	// ML routes
	ml := api.Group("/ml")
	ml.Get("/status", handlers.GetMLStatus)
	ml.Post("/train", middleware.AuthRequired(), handlers.TrainMLModels)

	// New ML service routes
	// Sentiment Analysis
	ml.Get("/sentiment/product/:id", middleware.AuthRequired(), handlers.GetProductSentiment)
	ml.Get("/sentiment/category/:category", middleware.AuthRequired(), handlers.GetCategorySentiment)
	ml.Get("/sentiment/insights", middleware.AuthRequired(), handlers.GetSentimentInsights)

	// Auto-Tagging
	ml.Get("/auto-tagging/suggest/:id", middleware.AuthRequired(), handlers.SuggestProductTags)
	ml.Post("/auto-tagging/auto-tag", middleware.AuthRequired(), handlers.AutoTagProducts)
	ml.Get("/auto-tagging/insights", middleware.AuthRequired(), handlers.GetTaggingInsights)

	// Smart Discounts
	ml.Get("/smart-discounts/suggest/product/:id", middleware.AuthRequired(), handlers.SuggestProductDiscount)
	ml.Get("/smart-discounts/suggest/category/:category", middleware.AuthRequired(), handlers.SuggestCategoryDiscounts)
	ml.Get("/smart-discounts/insights", middleware.AuthRequired(), handlers.GetDiscountInsights)

	// ML Services Management
	ml.Post("/initialize-services", middleware.AuthRequired(), handlers.InitializeMLServices)

	// Data enrichment routes
	// Favorites
	favorites := api.Group("/favorites", middleware.AuthRequired())
	favorites.Get("/", handlers.GetFavorites)
	favorites.Post("/", handlers.AddFavorite)
	favorites.Delete("/:product_id", handlers.RemoveFavorite)

	// Upvotes
	upvotes := api.Group("/upvotes")
	upvotes.Get("/:product_id", handlers.GetProductUpvotes)
	upvotes.Post("/", middleware.AuthRequired(), handlers.AddUpvote)
	upvotes.Delete("/:product_id", middleware.AuthRequired(), handlers.RemoveUpvote)

	// Comments
	comments := api.Group("/comments")
	comments.Get("/:product_id", handlers.GetProductComments)
	comments.Post("/", middleware.AuthRequired(), handlers.AddComment)
	comments.Put("/:comment_id", middleware.AuthRequired(), handlers.UpdateComment)
	comments.Delete("/:comment_id", middleware.AuthRequired(), handlers.DeleteComment)

	// Tags
	tags := api.Group("/tags")
	tags.Get("/", handlers.GetTags)
	tags.Post("/", middleware.AuthRequired(), handlers.CreateTag)
	tags.Get("/products/:product_id", handlers.GetProductTags)
	tags.Post("/products", middleware.AuthRequired(), handlers.AddProductTag)

	// Discounts
	discounts := api.Group("/discounts")
	discounts.Get("/active", handlers.GetActiveDiscounts)
	discounts.Post("/", middleware.AuthRequired(), handlers.CreateDiscount)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Route not found",
		})
	})

	// Get port from environment
	port := getEnv("PORT", "8080")

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("🔒 Security features enabled: Helmet, Rate Limiting, CORS")
	log.Printf("📊 Request timeout: 30s, Body limit: 4MB")
	log.Fatal(app.Listen(":" + port))
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
