package main

import (
	"context"
	"jooble-parser/internal/app"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := makeConfig()
	logger := makeLogger(cfg)
	htmlLoader := makeLoader(cfg, logger)
	jobsParser := makeParser(logger)
	dif := makeDiff(cfg, logger)
	signal := makeUpdateSignal(cfg, logger)

	defer logger.Sync()

	app := app.New(
		cfg,
		logger,
		htmlLoader,
		jobsParser,
		dif,
		signal)

	app.Run(gracefulShutDown())
}

func gracefulShutDown() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-c

		log.Println("services stopped by gracefulShutDown")
		cancel()
	}()

	return ctx
}
