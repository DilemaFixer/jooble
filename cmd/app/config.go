package main

import (
	"fmt"
	"jooble-parser/internal/config"
	"jooble-parser/internal/consts"
)

func makeConfig() *config.Config {
	cfg, err := config.Load(consts.ConfigPath)
	if err != nil {
		panic(fmt.Sprintf("Error loading config: %v", err))
	}
	return cfg
}
