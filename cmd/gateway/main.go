package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Raphel6969/api-gateway/internal/auth"
	"github.com/Raphel6969/api-gateway/internal/config"
	"github.com/Raphel6969/api-gateway/internal/router"
	"github.com/Raphel6969/api-gateway/middleware"
	"github.com/Raphel6969/api-gateway/pkg/logger"
)

func main() {
	// 1. Initialize Logger
	log := logger.New()
	log.Info("Starting API Gateway...")

	// Load Config
	cfg := config.LoadStaticConfig()

	// Setup a basic router
	rt, err := router.New(cfg)
	if err != nil {
		log.Error("Failed to initiate router ", "error", err)
		os.Exit(1)
	}

	addr := cfg.Port
	if !strings.HasSuffix(addr, ":") {
		addr = ":" + addr
	}

	// Initialize Authenticator
	authEngine := auth.NewAuthenticator(
		"super-secret-gateway-key-32-bytes", // Secret key
		map[string]string{
			"key-client-123": "user_api_partner_1", // Valid API Keys
		},
	)

	// Define public routes that do NOT require authentication
	publicPaths := []string{"/public"}

	pipeline := middleware.Chain(
		rt,
		middleware.Authenticate(*authEngine, publicPaths, log),
		middleware.Recovery(log),
		middleware.RequestID,
		middleware.Logging(log),
	)

	// Configure the HTTP Server
	srv := &http.Server{
		Addr:         addr,
		Handler:      pipeline,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server
	go func() {
		log.Info("Gateway Listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server Failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down server gracefully. . .")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
	}

	log.Info("Server exited properly")
}
