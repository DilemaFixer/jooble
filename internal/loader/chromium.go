package loader

import (
	"context"
	"fmt"
	"jooble-parser/internal/config"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

type (
	ChromeLoader struct {
		options        []chromedp.ExecAllocatorOption
		userDataFolder string
		logger         *zap.Logger
	}
)

func NewChromeLoader(cfg *config.ChromeConfig, logger *zap.Logger) HtmlLoader {
	return &ChromeLoader{
		options: []chromedp.ExecAllocatorOption{
			chromedp.ExecPath(cfg.ExePath),
			chromedp.UserDataDir(cfg.UserDataFolder),
			chromedp.Flag("headless", false),
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
		},
		logger:         logger,
		userDataFolder: cfg.UserDataFolder,
	}
}

func (loader *ChromeLoader) Load(url string, ctx context.Context) (string, error) {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), loader.options...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	var html string

	loader.logger.Debug("Loading", zap.String("url", url))
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),

		chromedp.WaitVisible(`div[data-test-name="_jobCard"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		return "", fmt.Errorf("chromedp error: %w", err)
	}
	_ = chromedp.Cancel(ctx)
	err = loader.clearUserData()
	return html, nil
}

func (loader *ChromeLoader) clearUserData() error {
	if err := os.RemoveAll(loader.userDataFolder); err != nil {
		return fmt.Errorf("error user data clearing %s: %w", loader.userDataFolder, err)
	}

	loader.logger.Debug("Clear user data folder", zap.String("path", loader.userDataFolder))
	return nil
}
