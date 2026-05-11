package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *slog.Logger

func InitLogger() {
	fileLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	// Write to both standard output and the file
	writer := io.MultiWriter(os.Stdout, fileLogger)

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	Log = slog.New(handler)
	slog.SetDefault(Log)
}
