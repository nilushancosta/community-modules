// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort             string
	ObserverAPIInternalURL string
	LogLevel               slog.Level
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	serverPort := getEnv("SERVER_PORT", "9098")
	observerAPIInternalURL := getEnv("OBSERVER_API_INTERNAL_URL", "")

	logLevel := slog.LevelInfo
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN", "WARNING":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}

	if observerAPIInternalURL == "" {
		return nil, fmt.Errorf("environment variable OBSERVER_API_INTERNAL_URL is required")
	}
	parsedURL, err := url.Parse(observerAPIInternalURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("OBSERVER_API_INTERNAL_URL must be a valid URL with scheme and host, got: %q", observerAPIInternalURL)
	}

	if _, err := strconv.Atoi(serverPort); err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	return &Config{
		ServerPort:             serverPort,
		ObserverAPIInternalURL: observerAPIInternalURL,
		LogLevel:               logLevel,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
