package main

import (
	"jooble-parser/internal/config"
	"jooble-parser/internal/loader"

	"go.uber.org/zap"
)

func makeLoader(cfg *config.Config, logger *zap.Logger) loader.HtmlLoader {
	return loader.NewChromeLoader(&cfg.Chrome, logger)
}
