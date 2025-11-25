package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juandiii/jetson-monitor/app"
	cfgpkg "github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/version"
)

func main() {
	log := logging.NewLogger()
	defer log.Sync()

	log.Infof("version=%s commit=%s", version.Version, version.Commit)

	cfg, err := cfgpkg.Load("config.yml", log)
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		os.Exit(1)
	}

	log.Debugf(
		"config: port=%d http_timeout=%ds per_attempt_timeout=%ds retry_max_tries=%d backoff_min_ms=%d backoff_max_ms=%d backoff_factor=%.2f backoff_jitter=%v urls=%d",
		cfg.Port, cfg.HTTPTimeoutSeconds, cfg.PerAttemptTimeoutSeconds, cfg.RetryMaxTries,
		cfg.BackoffMinMs, cfg.BackoffMaxMs, cfg.BackoffFactor, *cfg.BackoffJitter, len(cfg.Urls),
	)
	for i, u := range cfg.Urls {
		var expected string
		if u.StatusCode != nil {
			expected = fmt.Sprint(*u.StatusCode)
		} else {
			expected = "(2xx)"
		}
		log.Debugf("config.url[%d]=%s expected_status=%s scheduler=%s", i, u.URL, expected, u.Scheduler)
	}

	application := app.NewApp(cfg, log)
	serverErr := application.Start()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-sigCtx.Done():
		log.Infof("shutdown signal received, stopping...")
	case err := <-serverErr:
		if err != nil {
			log.Errorf("HTTP server stopped unexpectedly: %v", err)
		} else {
			log.Infof("HTTP server stopped")
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Errorf("error during shutdown: %v", err)
	}

	select {
	case err := <-serverErr:
		if err != nil {
			log.Errorf("server error: %v", err)
		}
	default:
	}

	<-shutdownCtx.Done()
	log.Infof("exiting")
}
