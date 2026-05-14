package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cache-server/internal/cache/service"
	"cache-server/internal/config"
	"cache-server/internal/http/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	// Load and validate configuration

	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create service
	srv := service.NewService(cfg)
	h := handlers.NewHandlers(srv)

	// Configure Fiber v3 with production settings
	app := fiber.New(fiber.Config{
		AppName:      "Production Cache Server",
		ServerHeader: "CacheServer",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    10 * 1024 * 1024, // 10MB max request size
	})

	// Setup routes
	api := app.Group("/api/v1")
	api.Post("/set", h.SetHandler)
	api.Get("/get/:key", h.GetHandler)
	api.Delete("/del/:key", h.DelHandler)
	api.Get("/metrics", h.MetricsHandler)
	app.Get("/health", h.HealthHandler)

	// Start janitor in background
	ctx, cancel := context.WithCancel(context.Background())
	go srv.Janitor(ctx)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Println("Server starting on :8080")
		if err := app.Listen(":1100"); err != nil {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal or server error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("Shutdown signal received")
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	}

	// Graceful shutdown
	log.Println("Shutting down gracefully...")

	// Stop janitor
	cancel()

	// Stop service (clock cleanup)
	srv.Stop()

	// Shutdown HTTP server with timeout
	if err := app.Shutdown(); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
