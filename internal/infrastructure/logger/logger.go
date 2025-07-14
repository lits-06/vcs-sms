package logger

import (
	"os"

	"github.com/lits-06/vcs-sms/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps zap logger
type Logger struct {
	*zap.SugaredLogger
}

// New creates a new logger instance
func New(cfg config.LoggerConfig) *Logger {
	// Create log directory if it doesn't exist
	if cfg.OutputPath != "" {
		os.MkdirAll(cfg.OutputPath[:len(cfg.OutputPath)-len("/app.log")], 0755)
	}

	// Configure log rotation
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.OutputPath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})

	// Parse log level
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	// Configure encoder
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "timestamp"
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		w,
		level,
	)

	logger := zap.New(core, zap.AddCaller())
	return &Logger{SugaredLogger: logger.Sugar()}
}
