package main

import (
	"log"
	"os"

	"bachelor_backend/database"
	"bachelor_backend/handlers"
	"bachelor_backend/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	// Initialize database connection
	database.Connect()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // React dev servers
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Session-ID"},
		AllowCredentials: true,
	}))

	// Health check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "bachelor_backend",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Authentication routes
	auth := api.Group("/auth")
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

	// Shopping cart routes (to be implemented)
	cart := api.Group("/cart", middleware.AuthRequired())
	cart.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Cart endpoints coming soon"})
	})

	// Order routes (to be implemented)
	orders := api.Group("/orders", middleware.AuthRequired())
	orders.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Order endpoints coming soon"})
	})

	// ML routes (to be implemented)
	ml := api.Group("/ml")
	ml.Get("/trends", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ML trends endpoint coming soon"})
	})

	// Chatbot routes (to be implemented)
	chat := api.Group("/chat")
	chat.Post("/message", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Chatbot endpoint coming soon"})
	})

	// Analytics routes (to be implemented)
	analytics := api.Group("/analytics", middleware.AuthRequired())
	analytics.Get("/dashboard", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Analytics dashboard coming soon"})
	})

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
