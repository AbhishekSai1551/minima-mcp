package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minima-global/minima-mcp/internal/audit"
	"github.com/minima-global/minima-mcp/internal/config"
	"github.com/minima-global/minima-mcp/internal/mcp"
	"github.com/minima-global/minima-mcp/internal/minima"
	"github.com/minima-global/minima-mcp/internal/ratelimit"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logLevel := parseLogLevel(cfg.Server.LogLevel)
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	logger.Info("starting minima-mcp server",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"auth_mode", cfg.Auth.Mode,
		"minima_data_url", cfg.Minima.DataURL(),
	)

	minimaClient := minima.NewClient(&cfg.Minima)

	var auditLogger *audit.Logger
	if cfg.Audit.Enabled {
		auditLogger, err = audit.NewLogger(cfg.Audit.LogDir, cfg.Audit.MaxAgeDays, cfg.Audit.Format)
		if err != nil {
			logger.Error("failed to create audit logger", "error", err)
			os.Exit(1)
		}
		defer auditLogger.Close()

		go func() {
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				if err := auditLogger.PurgeOldLogs(); err != nil {
					logger.Error("audit log purge failed", "error", err)
				}
			}
		}()
	} else {
		auditLogger = &audit.Logger{}
	}

	rateLimiter := ratelimit.NewRateLimiter(
		cfg.RateLimit.RequestsPerSecond,
		cfg.RateLimit.BurstSize,
		cfg.RateLimit.CleanupInterval,
	)

	mcpServer := mcp.NewMCPServer(minimaClient, auditLogger, rateLimiter, cfg.Auth.Mode, logger)

	mcp.RegisterNodeTools(mcpServer, minimaClient, logger)
	mcp.RegisterWalletTools(mcpServer, minimaClient, logger)
	mcp.RegisterContractTools(mcpServer, minimaClient, logger)
	mcp.RegisterTokenTools(mcpServer, minimaClient, logger)
	mcp.RegisterKeyTools(mcpServer, minimaClient, logger)
	mcp.RegisterMiniDAPPTools(mcpServer, minimaClient, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", mcpServer.HandleMCP)
	mux.HandleFunc("/mcp/", mcpServer.HandleMCP)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status, err := minimaClient.GetStatus(r.Context())
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"unhealthy","error":"%s"}`, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"minima": status,
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced shutdown", "error", err)
	}

	logger.Info("server stopped")
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
