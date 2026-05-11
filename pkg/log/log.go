package log

import (
	"log/slog"
	"os"
	"strings"
)

// SetLogger sets nginx-traefik-converter logger with desired log level.
func SetLogger(logLevel string) *slog.Logger {
	loggerOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     setLogLevel(logLevel),
	}

	stdLogger := slog.NewTextHandler(os.Stderr, loggerOpts)

	return slog.New(stdLogger)
}

func setLogLevel(logLevel string) slog.Level {
	switch strings.ToLower(logLevel) {
	case strings.ToLower(slog.LevelWarn.String()):
		return slog.LevelWarn
	case strings.ToLower(slog.LevelDebug.String()):
		return slog.LevelDebug
	case strings.ToLower(slog.LevelError.String()):
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
