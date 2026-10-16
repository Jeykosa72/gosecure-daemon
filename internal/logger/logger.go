package logger

import (
	"log/slog"
	"os"
	"strings"
)

// SetupLogger menginisialisasi structured logger global (JSON Handler untuk produksi/systemd).
func SetupLogger(levelStr string) *slog.Logger {
	var level slog.Level
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Menggunakan JSONHandler agar log mudah diparse oleh journalctl/Logstash
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	loggerInstance := slog.New(handler)

	// Mengatur sebagai logger standar global
	slog.SetDefault(loggerInstance)
	return loggerInstance
}