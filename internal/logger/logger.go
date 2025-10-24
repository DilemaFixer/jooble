package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"jooble-parser/internal/config"
)

func NewLogger(cfg config.LogConfig) (*zap.Logger, error) {
	level := getLogLevel(cfg.LogLevel)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	if cfg.LogToFile {
		if cfg.FilePath != "" {
			logDir := filepath.Dir(cfg.FilePath)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return nil, err
			}
		}

		fileWriter := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter),
			level,
		)

		cores = append(cores, fileCore)
	}

	consoleEncoderConfig := encoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	cores = append(cores, consoleCore)

	core := zapcore.NewTee(cores...)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger, nil
}

func getLogLevel(level int) zapcore.Level {
	switch level {
	case 0:
		return zapcore.DebugLevel
	case 1:
		return zapcore.InfoLevel
	case 2:
		return zapcore.WarnLevel
	case 3:
		return zapcore.ErrorLevel
	case 4:
		return zapcore.DPanicLevel
	case 5:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

func NewSugarLogger(cfg config.LogConfig) (*zap.SugaredLogger, error) {
	logger, err := NewLogger(cfg)
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
