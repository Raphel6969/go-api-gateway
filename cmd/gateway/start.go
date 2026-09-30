package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Raphel6969/api-gateway/internal/auth"
	"github.com/Raphel6969/api-gateway/internal/config"
	"github.com/Raphel6969/api-gateway/internal/config/cache"
	"github.com/Raphel6969/api-gateway/internal/metrics"
	"github.com/Raphel6969/api-gateway/internal/middleware"
	"github.com/Raphel6969/api-gateway/internal/ratelimit"
	"github.com/Raphel6969/api-gateway/internal/router"
	"github.com/Raphel6969/api-gateway/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var cfgFile string

// DynamicHandler holds our middleware pipeline safely in memory
type DynamicHandler struct {
	pipeline atomic.Value
}

func (h *DynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Load the current pipeline and pass the request to it
	handler := h.pipeline.Load().(http.Handler)
	handler.ServeHTTP(w, r)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the API Gateway server",
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.New()
		log.Info("Starting API Gateway...", "config", cfgFile)

		// Initialize Prometheus metrics collectors once
		metricsRegistry := metrics.New()

		// 1. Build the initial pipeline
		initialPipeline, initialCfg, err := buildPipeline(cfgFile, metricsRegistry, log)
		if err != nil {
			log.Error("Failed to build initial pipeline", "error", err)
			os.Exit(1)
		}

		// 2. Wrap it in our thread-safe DynamicHandler
		dynamicHandler := &DynamicHandler{}
		dynamicHandler.pipeline.Store(initialPipeline)

		// 3. Start the background config watcher
		go watchConfig(cfgFile, metricsRegistry, dynamicHandler, log)

		// 4. Start the HTTP Server
		addr := initialCfg.Server.Port
		if !strings.HasPrefix(addr, ":") {
			addr = ":" + addr
		}

		srv := &http.Server{
			Addr:         addr,
			Handler:      dynamicHandler, // Use the dynamic handler!
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		go func() {
			log.Info("Gateway Listening", "port", addr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error("Server Failed", "error", err)
				os.Exit(1)
			}
		}()

		// Graceful Shutdown
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Info("Shutting down server gracefully...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Server forced to shutdown", "error", err)
		}
		log.Info("Server exited properly")
	},
}

// buildPipeline reads the YAML and constructs the router and middlewares
func buildPipeline(filename string, m *metrics.Metrics, log *slog.Logger) (http.Handler, *config.Config, error) {
	cfg, err := config.LoadConfig(filename)
	if err != nil {
		return nil, nil, err
	}

	rt, err := router.New(cfg, log)
	if err != nil {
		return nil, nil, err
	}

	publicPaths := []string{"/metrics", "/metrics/"}
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
		middleware.Metrics(m),
		middleware.Logging(log),
		middleware.Authenticate(*authEngine, publicPaths, log),
		middleware.RateLimit(limiter, log),
		middleware.CacheMiddleware(memCache, cacheTTL, m, log),
	)

	return pipeline, cfg, nil
}

// watchConfig listens for file saves and atomically hot-swaps the pipeline
func watchConfig(filename string, m *metrics.Metrics, dynamicHandler *DynamicHandler, log *slog.Logger) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Failed to start config watcher", "error", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(filename)
	if err != nil {
		log.Error("Failed to watch config file", "error", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// If the file was written to (saved)
			if event.Has(fsnotify.Write) {
				log.Info("Config change detected! Reloading Gateway...")

				// Add a tiny delay because some text editors write in chunks
				time.Sleep(100 * time.Millisecond)

				newPipeline, _, err := buildPipeline(filename, m, log)
				if err != nil {
					log.Error("Failed to reload config (keeping old config)", "error", err)
				} else {
					dynamicHandler.pipeline.Store(newPipeline)
					log.Info("✅ Gateway reloaded successfully without downtime!")
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("Watcher error", "error", err)
		}
	}
}

func init() {
	startCmd.Flags().StringVarP(&cfgFile, "config", "c", "gateway.yaml", "config file (default is gateway.yaml)")
	rootCmd.AddCommand(startCmd)
}
