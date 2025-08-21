package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/repositories"
	"backend/internal/services"
	"backend/pkg/redis"
)

// main is the application entry point
// Sets up the server, dependencies, and graceful shutdown handling
func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Redis client
	redisClient, err := redis.NewClient(cfg.Redis.Config)
	if err != nil {
		log.Fatalf("Error 1001: Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize Gin router
	router := setupRouter(cfg, redisClient)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error 1002: Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a 30-second timeout to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error 1003: Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// setupRouter configures and returns the Gin router with all routes and middleware
// Sets up health checks, API routes, and middleware stack with dependency injection
func setupRouter(cfg *config.Config, redisClient *redis.Client) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add CORS middleware if enabled
	if cfg.Security.EnableCORS {
		router.Use(corsMiddleware(cfg.Security.AllowedOrigins))
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(redisClient)
	taskRepo := repositories.NewTaskRepository(redisClient)

	// Initialize services
	userService := services.NewUserService(userRepo)
	taskService := services.NewTaskService(taskRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(userRepo)

	// Health check endpoint
	router.GET("/health", healthCheckHandler(redisClient))

	// API version 1 routes group
	v1 := router.Group("/api/v1")
	{
		// Auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(authMiddleware)
		{
			// Auth routes that require authentication
			protected.GET("/auth/me", authHandler.Me)
			// Task routes
			protected.GET("/tasks", taskHandler.ListTasks)
			protected.POST("/tasks", taskHandler.CreateTask)
			protected.GET("/tasks/:id", taskHandler.GetTask)
			protected.PUT("/tasks/:id/complete", taskHandler.UpdateTaskCompletion)
			protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
			protected.POST("/tasks/:id/restore", taskHandler.RestoreTask)

			// Category routes
			protected.GET("/categories", taskHandler.GetCategories)
			protected.PUT("/categories/:categoryName", taskHandler.RenameCategory)
			protected.DELETE("/categories/:categoryName", taskHandler.DeleteCategory)
		}
	}

	return router
}

// healthCheckHandler returns a handler for health check endpoints
// Verifies that the application and its dependencies are running properly
func healthCheckHandler(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Check Redis connection
		if err := redisClient.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "Redis connection failed",
				"code":   "1004",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	}
}

// corsMiddleware returns a CORS middleware function
// Configures Cross-Origin Resource Sharing for frontend integration
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}