package log

import (
	"log/slog"
	"testing"
)

func TestSetLogger(t *testing.T) {
	if logger := SetLogger("debug"); logger == nil {
		t.Fatal("expected logger")
	}
}

func TestSetLogLevel(t *testing.T) {
	tests := map[string]slog.Level{
		"warn":  slog.LevelWarn,
		"DEBUG": slog.LevelDebug,
		"error": slog.LevelError,
		"info":  slog.LevelInfo,
		"bogus": slog.LevelInfo,
	}

	for input, expected := range tests {
		if actual := setLogLevel(input); actual != expected {
			t.Fatalf("setLogLevel(%q) = %s, want %s", input, actual, expected)
		}
	}
}
