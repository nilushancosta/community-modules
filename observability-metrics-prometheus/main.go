// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	app "github.com/openchoreo/community-modules/observability-metrics-prometheus/internal"
	"github.com/openchoreo/community-modules/observability-metrics-prometheus/internal/observer"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		logger.Error("Failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	logger.Info("Configurations loaded from environment variables successfully",
		slog.String("Log Level", cfg.LogLevel.String()),
		slog.String("Server Port", cfg.ServerPort),
	)

	observerClient := observer.NewClient(cfg.ObserverAPIInternalURL)
	metricsHandler := app.NewMetricsHandler(observerClient, logger)
	srv := app.NewServer(cfg.ServerPort, metricsHandler, logger)

	serverErrCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			serverErrCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	exitCode := 0
	select {
	case <-quit:
	case err := <-serverErrCh:
		logger.Error("Server error", slog.Any("error", err))
		exitCode = 1
	}

	logger.Info("Shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error during shutdown", slog.Any("error", err))
		exitCode = 1
	}

	logger.Info("Server stopped")
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
