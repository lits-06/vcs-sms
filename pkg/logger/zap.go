package logger

import (
	"os"

	"github.com/lits-06/vcs-sms/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	logger *zap.Logger
}

type Config struct {
	Level      string `yaml:"level" json:"level"`             // debug, info, warn, error
	LogFile    string `yaml:"log_file" json:"log_file"`       // đường dẫn file log
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // MB
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // số file backup
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // ngày
	Compress   bool   `yaml:"compress" json:"compress"`       // nén file cũ
}

func NewZapLogger() (*zapLogger, error) {
	logDir := "../../logs"

	// Cấu hình log rotation
	logRotator := &lumberjack.Logger{
		Filename:   logDir + "/" + config.LOG_FILE,
		MaxSize:    config.LOG_MAX_AGE,     // MB
		MaxBackups: config.LOG_MAX_BACKUPS, // số file backup
		MaxAge:     config.LOG_MAX_AGE,     // ngày
		Compress:   config.LOG_COMPRESSION, // nén file cũ
	}

	// Cấu hình encoder cho JSON format
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Parse log level
	level, err := zapcore.ParseLevel(config.LOG_LEVEL)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Tạo core cho file output
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logRotator),
		level,
	)

	// Tạo core cho console output (development)
	consoleConfig := encoderConfig
	consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Combine cả hai cores
	core := zapcore.NewTee(fileCore, consoleCore)

	// Tạo logger với caller info
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &zapLogger{logger: logger}, nil
}

func (l *zapLogger) Info(message string, args ...interface{}) {
	l.logger.Info(message, l.argsToFields(args...)...)
}

func (l *zapLogger) Debug(message string, args ...interface{}) {
	l.logger.Debug(message, l.argsToFields(args...)...)
}

func (l *zapLogger) Warn(message string, args ...interface{}) {
	l.logger.Warn(message, l.argsToFields(args...)...)
}

func (l *zapLogger) Error(message string, args ...interface{}) {
	l.logger.Error(message, l.argsToFields(args...)...)
}

func (l *zapLogger) Fatal(message string, args ...interface{}) {
	l.logger.Fatal(message, l.argsToFields(args...)...)
}

func (l *zapLogger) Panic(message string, args ...interface{}) {
	l.logger.Panic(message, l.argsToFields(args...)...)
}

func (l *zapLogger) With(key string, value interface{}) Logger {
	newLogger := l.logger.With(zap.Any(key, value))
	return &zapLogger{logger: newLogger}
}

// argsToFields chuyển đổi args thành zap fields
func (l *zapLogger) argsToFields(args ...interface{}) []zap.Field {
	if len(args)%2 != 0 {
		// Nếu số args lẻ, thêm một giá trị nil
		args = append(args, nil)
	}

	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, args[i+1]))
	}

	return fields
}

// Close đóng logger
func (l *zapLogger) Close() error {
	return l.logger.Sync()
}
