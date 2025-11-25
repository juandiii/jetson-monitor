package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/juandiii/jetson-monitor/config"
)

type nopLogger struct{}

func (n *nopLogger) Debugf(format string, args ...interface{}) {}
func (n *nopLogger) Infof(format string, args ...interface{})  {}
func (n *nopLogger) Errorf(format string, args ...interface{}) {}
func (n *nopLogger) Debug(args ...interface{})                 {}
func (n *nopLogger) Info(args ...interface{})                  {}
func (n *nopLogger) Error(args ...interface{})                 {}
func (n *nopLogger) Sync() error                               { return nil }

// spyLogger captures Infof/Errorf calls for assertions.
type spyLogger struct {
	infos []string
	errs  []string
}

func (s *spyLogger) Debugf(format string, args ...interface{}) {}
func (s *spyLogger) Infof(format string, args ...interface{})  { s.infos = append(s.infos, fmt.Sprintf(format, args...)) }
func (s *spyLogger) Errorf(format string, args ...interface{}) { s.errs = append(s.errs, fmt.Sprintf(format, args...)) }
func (s *spyLogger) Debug(args ...interface{})                 {}
func (s *spyLogger) Info(args ...interface{})                  {}
func (s *spyLogger) Error(args ...interface{})                 {}
func (s *spyLogger) Sync() error                               { return nil }

func TestAppStartShutdown(t *testing.T) {
	tmp := t.TempDir()
	stateFile := filepath.Join(tmp, "state.json")

	cfg := &config.ConfigJetson{
		Port:      0, // let OS pick a free port
		StateFile: stateFile,
		Urls:      []config.URL{},
	}

	log := &nopLogger{}

	a := NewApp(cfg, log)

	// Start the app
	errCh := a.Start()

	// give server a short moment to start
	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server returned error immediately: %v", err)
		}
	default:
		// no immediate error, good
	}

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := a.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}

	// ensure state file path's directory exists (store may create it)
	if _, err := os.Stat(tmp); err != nil {
		t.Fatalf("expected temp dir to exist: %v", err)
	}
}

func TestNewApp_AssignsLoggerAndStore(t *testing.T) {
	tmp := t.TempDir()
	stateFile := filepath.Join(tmp, "state.json")

	cfg := &config.ConfigJetson{
		Port:      0,
		StateFile: stateFile,
		Urls:      []config.URL{},
	}

	spy := &spyLogger{}

	a := NewApp(cfg, spy)

	if a == nil {
		t.Fatalf("expected NewApp to return non-nil App")
	}
	if cfg.Logger == nil {
		t.Fatalf("expected cfg.Logger to be assigned by NewApp")
	}
	if a.store == nil {
		t.Fatalf("expected store to be created in NewApp")
	}
}

func TestStart_InvalidCron_LogsError(t *testing.T) {
	tmp := t.TempDir()
	stateFile := filepath.Join(tmp, "state.json")

	cfg := &config.ConfigJetson{
		Port:      0,
		StateFile: stateFile,
		Urls: []config.URL{
			{URL: "http://example.local", Scheduler: "@every invalid"},
		},
	}

	spy := &spyLogger{}

	a := NewApp(cfg, spy)

	errCh := a.Start()
	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server returned error immediately: %v", err)
		}
	default:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := a.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}

	if len(spy.errs) == 0 {
		t.Fatalf("expected spy logger to record an error when scheduling invalid cron")
	}
}

func TestStart_ValidCron_AddsJob(t *testing.T) {
	tmp := t.TempDir()
	stateFile := filepath.Join(tmp, "state.json")

	cfg := &config.ConfigJetson{
		Port:      0,
		StateFile: stateFile,
		Urls: []config.URL{
			{URL: "http://example.local", Scheduler: "@every 1s"},
		},
	}

	spy := &spyLogger{}
	a := NewApp(cfg, spy)

	errCh := a.Start()
	// give server and cron a moment
	time.Sleep(200 * time.Millisecond)

	// ensure server started without immediate error
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server returned error immediately: %v", err)
		}
	default:
	}

	// Check that cron has at least one entry
	if a.cronx == nil {
		t.Fatalf("expected cron to be initialized")
	}
	entries := a.cronx.Entries()
	if len(entries) == 0 {
		t.Fatalf("expected cron to have at least one entry")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := a.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}
}
