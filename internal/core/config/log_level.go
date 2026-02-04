package config

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	// LevelSilent represents a silent log level (no output)
	// Using a very high value ensures nothing gets logged
	LevelSilent slog.Level = slog.LevelError + 100
)

// ParseLogLevel converts a string log level to slog.Level.
// Supported values: "silent", "debug", "info", "warn", "error" (case-insensitive).
// Returns slog.LevelInfo if the input is empty or invalid.
func ParseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "silent":
		return LevelSilent
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default to Info level
	}
}

// ConfigureLogger sets up the global slog logger based on the provided level.
// For silent mode, it configures a handler that discards all output.
// For other levels, it configures a text handler writing to stderr.
func ConfigureLogger(userLevel string) {
	var handler slog.Handler
	level := ParseLogLevel(userLevel)

	if level == LevelSilent {
		// Create a handler that writes to io.Discard for silent mode
		handler = slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})
	} else {
		// Create a normal text handler writing to stderr
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
	}

	slog.SetDefault(slog.New(handler))
}
