package main

import (
	"jooble-parser/internal/parser"
	"jooble-parser/internal/parser/setters"

	"go.uber.org/zap"
)

func makeParser(logger *zap.Logger) *parser.JobParser {
	return parser.NewJobParser(logger, setters.AllSetters...)
}
