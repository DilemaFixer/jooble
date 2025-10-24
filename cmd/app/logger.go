package main

import (
	"fmt"
	"jooble-parser/internal/config"
	"jooble-parser/internal/logger"

	"go.uber.org/zap"
)

func makeLogger(cfg *config.Config) *zap.Logger {
	logger, err := logger.NewLogger(cfg.Log)
	if err != nil {
		panic(fmt.Sprintf("Error initializing logger: %v", err))
	}
	return logger
}
