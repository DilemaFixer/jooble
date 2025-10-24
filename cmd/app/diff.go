package main

import (
	"fmt"
	"jooble-parser/internal/config"
	"jooble-parser/internal/differ"
	"jooble-parser/internal/service"

	"go.uber.org/zap"
)

func makeDiff(cfg *config.Config, logger *zap.Logger) differ.Differ {
	dbCfg := cfg.DB
	jobService, err := service.NewSqliteRepoService(
		dbCfg.Path,
		dbCfg.Limit,
		uint(dbCfg.ClearingStep),
		logger)
	if err != nil {
		panic(fmt.Sprintf("Error creating job service: %v", err))
	}

	dif := differ.NewDefaultDiffer(jobService)
	return dif
}
