package helpers

import (
	"log/slog"
	"os"
	"strings"
)

const (
	TextType = "TEXT"
	JsonType = "JSON"
)

func InitializeLogger() *slog.Logger {
	level := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	opts := &slog.HandlerOptions{Level: getLogLevel(level)}

	logType := getLogType(strings.ToUpper(os.Getenv("LOG_TYPE")))

	var logger *slog.Logger

	if logType == JsonType {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	logger.Info("Logger Initialized", "Type", logType)
	return logger
}

func getLogType(logType string) string {
	switch logType {
	case JsonType:
		return JsonType
	default:
		return TextType
	}
}
func getLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
