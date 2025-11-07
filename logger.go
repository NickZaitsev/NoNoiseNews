package main

import (
	"log/slog"
	"os"
)

// initLogger initializes the global structured logger.
func initLogger() {
	// Create a JSON handler for production-ready structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	
	// Create and set the global logger
	slog.SetDefault(slog.New(handler))
}

// Log functions for convenience

// LogInfo logs an info message.
func LogInfo(msg string, args ...any) {
	slog.Info(msg, args...)
}

// LogError logs an error message.
func LogError(msg string, err error, args ...any) {
	slog.Error(msg, append(args, "error", err)...)
}

// LogWarn logs a warning message.
func LogWarn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// LogDebug logs a debug message.
func LogDebug(msg string, args ...any) {
	slog.Debug(msg, args...)
}