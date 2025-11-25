package app

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3"

	"github.com/juandiii/jetson-monitor/api"
	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"
	"github.com/juandiii/jetson-monitor/notification/slack"
	"github.com/juandiii/jetson-monitor/notification/telegram"
	"github.com/juandiii/jetson-monitor/scheduler"
)

type App struct {
	cfg       *config.ConfigJetson
	logger    logging.Logger
	store     scheduler.StateStore
	cronx     *cron.Cron
	http      *fiber.App
	serverErr chan error
}

func NewApp(cfg *config.ConfigJetson, log logging.Logger) *App {
	if cfg.Logger == nil {
		cfg.Logger = log
	}

	store := scheduler.NewFileStateStore(cfg.StateFile, log)

	return &App{
		cfg:    cfg,
		logger: log,
		store:  store,
	}
}

func (a *App) Start() <-chan error {

	tg := a.cfg.EffectiveTelegram()
	sl := a.cfg.EffectiveSlack()

	providers := []notification.CommandProvider{
		slack.New(sl, a.cfg.Logger, nil),
		telegram.New(tg, a.cfg.Logger, nil),
	}

	a.cronx = cron.New()
	for _, u := range a.cfg.Urls {
		if _, err := a.cronx.AddJob(u.Scheduler, scheduler.New(u, a.cfg, a.store, providers)); err != nil {
			if a.logger != nil {
				a.logger.Errorf("failed to schedule job for %s: %v", u.URL, err)
			}
			continue
		}
	}
	a.cronx.Start()
	if a.logger != nil {
		a.logger.Infof("cron started with %d jobs", len(a.cfg.Urls))
	}

	a.http = fiber.New(fiber.Config{DisableStartupMessage: true})
	api.InitializeRoute(a.http)

	a.serverErr = make(chan error, 1)
	go func() {
		a.serverErr <- a.http.Listen(fmt.Sprintf(":%d", a.cfg.Port))
	}()

	if a.logger != nil {
		a.logger.Infof("HTTP start :: listening port: %d", a.cfg.Port)
	}

	return a.serverErr
}

func (a *App) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		if a.http != nil {
			if err := a.http.Shutdown(); err != nil && a.logger != nil {
				a.logger.Errorf("error shutting down HTTP server: %v", err)
			}
		}
		close(done)
	}()

	select {
	case <-done:
		if a.logger != nil {
			a.logger.Infof("HTTP server shutdown complete")
		}
	case <-ctx.Done():
		if a.logger != nil {
			a.logger.Errorf("shutdown deadline reached: %v", ctx.Err())
		}
		return ctx.Err()
	}

	if a.cronx != nil {
		a.cronx.Stop()
	}
	if a.logger != nil {
		a.logger.Infof("cron jobs stopped")
	}

	return nil
}
