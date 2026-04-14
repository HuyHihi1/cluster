package observability

import (
	"log/slog"
	"os"
)

// InitLogger sets up a structured JSON logger using the standard log/slog package.
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}
