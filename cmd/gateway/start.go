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
	"github.com/Raphel6969/api-gateway/internal/cache"
	"github.com/Raphel6969/api-gateway/internal/config"
	"github.com/Raphel6969/api-gateway/internal/ratelimit"
	"github.com/Raphel6969/api-gateway/internal/router"
	"github.com/Raphel6969/api-gateway/middleware"
	"github.com/Raphel6969/api-gateway/pkg/logger"
	"github.com/spf13/cobra"
)

var cfgFile string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the API Gateway server",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Initialize Logger
		log := logger.New()
		log.Info("Starting API Gateway...")

		// Load Config
		cfg, err := config.LoadConfig("gateway.yaml")
		if err != nil {
			log.Error("Failed to load configuration", "error", err)
			os.Exit(1)
		}

		// Setup a basic router
		rt, err := router.New(cfg, log)
		if err != nil {
			log.Error("Failed to initiate router ", "error", err)
			os.Exit(1)
		}

		var publicPaths []string
		for _, r := range cfg.Routes {
			if r.Public {
				publicPaths = append(publicPaths, r.Path)
			}
		}

		authEngine := auth.NewAuthenticator(cfg.Auth.JWTSecret, cfg.Auth.APIKeys)
		limiter := ratelimit.NewRateLimiter(cfg.RateLimit.Capacity, cfg.RateLimit.RefillRate)
		memCache := cache.NewMemoryCache()
		cacheTTL := time.Duration(cfg.Cache.TTLSeconds) * time.Second

		pipeline := middleware.Chain(
			rt,
			middleware.Recovery(log),
			middleware.RequestID,
			middleware.Logging(log),
			middleware.Authenticate(*authEngine, publicPaths, log),
			middleware.RateLimit(limiter, log),
			middleware.CacheMiddleware(memCache, cacheTTL, log),
		)

		addr := cfg.Server.Port
		if !strings.HasPrefix(addr, ":") {
			addr = ":" + addr
		}
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
			log.Info("Gateway Listening", "port", cfg.Server.Port)
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
	},
}

func init() {
	startCmd.Flags().StringVarP(&cfgFile, "config", "c", "gateway.yaml", "config file (default is gateway.yaml")
	rootCmd.AddCommand(startCmd)
}
