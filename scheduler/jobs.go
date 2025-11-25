package scheduler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"

	"github.com/juandiii/jetson-monitor/sdk"
	"github.com/juandiii/jetson-monitor/version"
)

const defaultCooldown = time.Hour

var defaultUA = fmt.Sprintf("server-check/%s (os=%s; arch=%s; go=%s)", version.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())

type Scheduler struct {
	conf       config.URL
	parentConf *config.ConfigJetson

	logger    logging.Logger
	store     StateStore
	sdk       *sdk.Client
	providers []notification.CommandProvider

	cooldown time.Duration
}

func New(c config.URL, cfg *config.ConfigJetson, store StateStore, providers []notification.CommandProvider) cron.Job {
	statePath := cfg.StateFile
	if statePath == "" {
		statePath = "jetson_notified_state.json"
	}

	httpClient := standardHTTPClient(perAttemptTimeout(cfg))
	client, err := sdk.New(httpClient,
		sdk.WithRetryAndBackoffs(sdk.RetryConfig{
			RetryMax:     cfg.RetryMaxTries,
			RetryWaitMin: sdk.PtrTo(float64(cfg.BackoffMinMs) / 1000.0),
			RetryWaitMax: sdk.PtrTo(float64(cfg.BackoffMaxMs) / 1000.0),
		}),
		sdk.SetUserAgent(defaultUA),
	)
	if err != nil && cfg.Logger != nil {
		cfg.Logger.Errorf("sdk init error: %v", err)
	}

	cooldown := defaultCooldown
	if cfg.NotificationCooldownSeconds > 0 {
		cooldown = time.Duration(cfg.NotificationCooldownSeconds) * time.Second
	}

	return &Scheduler{
		conf:       c,
		parentConf: cfg,
		logger:     cfg.Logger,
		store:      store,
		sdk:        client,
		providers:  providers,
		cooldown:   cooldown,
	}
}

func (s *Scheduler) Run() {
	url := s.conf.URL

	ctx, cancel := context.WithTimeout(context.Background(), s.totalTimeout())
	defer cancel()

	req, err := s.sdk.NewRequest(ctx, http.MethodGet, url, nil)

	if err != nil {
		s.notifyDown(url, fmt.Errorf("build request: %w", err))
		return
	}

	resp, err := s.sdk.Do(ctx, req, io.Discard)

	var got int
	switch {
	case err == nil && resp != nil:
		got = resp.StatusCode

	case err != nil:
		var er *sdk.ErrorResponse
		if errors.As(err, &er) && er.Response != nil {
			got = er.Response.StatusCode
		} else {
			s.notifyDown(s.conf.URL, err)
			return
		}
	}

	if s.conf.StatusCode != nil {
		want := *s.conf.StatusCode
		if got == want {
			if s.store.IsCoolingDown(s.conf.URL, s.cooldown) {
				s.notifyRecovery(s.conf.URL)
			}
			_ = s.store.Clear(s.conf.URL)
			return
		}
		s.notifyDown(s.conf.URL, fmt.Errorf("unexpected status: got %d, want %d", got, want))
		return
	}

	if got < 200 || got >= 300 {
		s.notifyDown(s.conf.URL, fmt.Errorf("non-2xx status: %d", got))
		return
	}

	if s.store.IsCoolingDown(s.conf.URL, s.cooldown) {
		s.notifyRecovery(s.conf.URL)
	}
	_ = s.store.Clear(s.conf.URL)
}

func (s *Scheduler) notifyDown(url string, err error) {
	if s.store.IsCoolingDown(url, s.cooldown) {
		if s.logger != nil {
			s.logger.Debugf("already notified about %s; skipping duplicate: %v", url, err)
		}
		return
	}
	for _, p := range s.providers {
		if p != nil {
			if e := p.SendMessage(&notification.Message{
				Text: fmt.Sprintf("The server %s is down: %v", url, err),
			}); e != nil && s.logger != nil {
				s.logger.Errorf("failed sending notification: %v", e)
			}
		}

	}
	_ = s.store.Touch(url)
}

func (s *Scheduler) notifyRecovery(url string) {
	for _, p := range s.providers {
		if p != nil {
			if e := p.SendMessage(&notification.Message{
				Text: fmt.Sprintf("The server %s is back up", url),
			}); e != nil && s.logger != nil {
				s.logger.Errorf("failed sending recovery notification: %v", e)
			}
		}

	}
}

// helpers
func perAttemptTimeout(conf *config.ConfigJetson) time.Duration {
	if conf.PerAttemptTimeoutSeconds > 0 {
		return time.Duration(conf.PerAttemptTimeoutSeconds) * time.Second
	}
	if conf.HTTPTimeoutSeconds > 0 {
		return time.Duration(conf.HTTPTimeoutSeconds) * time.Second
	}
	return 10 * time.Second
}

func (s *Scheduler) totalTimeout() time.Duration {
	perAttempt := perAttemptTimeout(s.parentConf)
	r := s.parentConf.RetryMaxTries
	if r < 1 {
		r = 1
	}
	estimatedBackoff := time.Duration(s.parentConf.BackoffMaxMs) * time.Millisecond * time.Duration(r)
	return perAttempt*time.Duration(r) + estimatedBackoff + 5*time.Second
}

func standardHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}
