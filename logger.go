package nselib

import (
	"log/slog"
	"os"
)

// EnableLogging configures console logging at the given level.
func EnableLogging(level slog.Level) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout,
		&slog.HandlerOptions{Level: level})))
}
