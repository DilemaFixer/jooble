package main

import (
	"jooble-parser/internal/config"
	"jooble-parser/internal/signal"

	"go.uber.org/zap"
)

func makeUpdateSignal(cfg *config.Config, logger *zap.Logger) signal.UpdateSignal {
	return signal.NewBotUpdateSignal(cfg, logger)
}
