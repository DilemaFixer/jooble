package app

import (
	"context"
	"jooble-parser/internal/config"
	"jooble-parser/internal/differ"
	downloader "jooble-parser/internal/loader"
	htmlParser "jooble-parser/internal/parser"
	"jooble-parser/internal/signal"
	"time"

	"go.uber.org/zap"
)

type App struct {
	cfg    *config.Config
	logger *zap.Logger
	url    string

	loader downloader.HtmlLoader
	parser *htmlParser.JobParser
	differ differ.Differ
	signal signal.UpdateSignal
}

func New(cfg *config.Config,
	logger *zap.Logger,
	loader downloader.HtmlLoader,
	parser *htmlParser.JobParser,
	differ differ.Differ,
	sign signal.UpdateSignal) *App {

	return &App{
		cfg:    cfg,
		logger: logger,
		loader: loader,
		parser: parser,
		url:    cfg.Parsing.Url,
		differ: differ,
		signal: sign,
	}
}

func (app *App) Run(ctx context.Context) {
	loader := app.loader
	parser := app.parser
	differ := app.differ
	signal := app.signal
	logger := app.logger

	for {
		html, err := loader.Load(app.url, ctx)
		if err != nil {
			logger.Error("loader error", zap.Error(err))
			continue
		}

		jobs, err := parser.Parse(html)
		if err != nil {
			logger.Error("parser error", zap.Error(err))
			continue
		}

		new, err := differ.Check(jobs)
		if err != nil {
			logger.Error("differ error", zap.Error(err))
			continue
		}

		if err := signal.Signal(new); err != nil {
			logger.Error("update signal error", zap.Error(err))
		}

		app.Sleep()
	}
}

func (app *App) Sleep() {
	time.Sleep(time.Second * app.cfg.ParsingConfig.Delay)
}
