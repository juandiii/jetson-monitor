package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/goccy/go-yaml"
	"github.com/robfig/cron/v3"

	"github.com/juandiii/jetson-monitor/logging"
)

type TelegramCfg struct {
	Token  string `yaml:"token"`
	ChatID int64  `yaml:"chat_id"`
}
type SlackCfg struct {
	Webhook string `yaml:"webhook"`
	Token   string `yaml:"token,omitempty"`
}
type NotifCfg struct {
	Telegram *TelegramCfg `yaml:"telegram,omitempty"`
	Slack    *SlackCfg    `yaml:"slack,omitempty"`
}

type URL struct {
	URL        string `yaml:"url"`
	StatusCode *int   `yaml:"status_code"`
	Scheduler  string `yaml:"scheduler"`
}

type ConfigJetson struct {
	Port                        int     `yaml:"port"`
	StateFile                   string  `yaml:"state_file"`
	HTTPTimeoutSeconds          int     `yaml:"http_timeout_seconds"`
	PerAttemptTimeoutSeconds    int     `yaml:"per_attempt_timeout_seconds"`
	RetryMaxTries               int     `yaml:"retry_max_tries"`
	BackoffMinMs                int     `yaml:"backoff_min_ms"`
	BackoffMaxMs                int     `yaml:"backoff_max_ms"`
	BackoffFactor               float64 `yaml:"backoff_factor"`
	BackoffJitter               *bool   `yaml:"backoff_jitter"`
	NotificationCooldownSeconds int     `yaml:"notification_cooldown_seconds"`
	UserAgent                   string  `yaml:"user_agent"`

	Notifications *NotifCfg `yaml:"notifications,omitempty"`

	Urls   []URL          `yaml:"urls"`
	Logger logging.Logger `yaml:"-"`
}

func Load(path string, logger logging.Logger) (*ConfigJetson, error) {
	if path == "" {
		path = "config.yml"
	}
	if err := validateFile(path); err != nil {
		if logger != nil {
			logger.Error("failed load config.yml: ", err)
		}
		return nil, err
	}

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := &ConfigJetson{Logger: logger}
	if err := yaml.NewDecoder(f).Decode(c); err != nil {
		return nil, err
	}

	applyDefaults(c)
	applyEnvOverrides(c)

	if err := c.Validate(); err != nil {
		return nil, err
	}

	if c.StateFile != "" && !filepath.IsAbs(c.StateFile) {
		if abs, err := filepath.Abs(c.StateFile); err == nil {
			c.StateFile = abs
		}
	}
	return c, nil
}

func applyDefaults(c *ConfigJetson) {
	if c.Port == 0 {
		c.Port = 38080
	}
	if c.StateFile == "" {
		c.StateFile = filepath.Join(os.TempDir(), "jetson_notified_state.json")
	}
	if c.HTTPTimeoutSeconds == 0 {
		c.HTTPTimeoutSeconds = 30
	}

	if c.RetryMaxTries == 0 {
		c.RetryMaxTries = 3
	}
	if c.BackoffMinMs == 0 {
		c.BackoffMinMs = 500
	}
	if c.BackoffMaxMs == 0 {
		c.BackoffMaxMs = 10000
	}
	if c.BackoffFactor == 0 {
		c.BackoffFactor = 2.0
	}
	if c.BackoffJitter == nil {
		b := true
		c.BackoffJitter = &b
	}
	if c.NotificationCooldownSeconds == 0 {
		c.NotificationCooldownSeconds = 3600 // 1h
	}
}

func applyEnvOverrides(c *ConfigJetson) {
	if v := os.Getenv("JETSON_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Port = n
		}
	}
	if v := os.Getenv("JETSON_STATE_FILE"); v != "" {
		c.StateFile = v
	}
	if v := os.Getenv("JETSON_HTTP_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.HTTPTimeoutSeconds = n
		}
	}
	if v := os.Getenv("JETSON_PER_ATTEMPT_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.PerAttemptTimeoutSeconds = n
		}
	}
	if v := os.Getenv("JETSON_RETRY_MAX_TRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.RetryMaxTries = n
		}
	}
	if v := os.Getenv("JETSON_BACKOFF_MIN_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.BackoffMinMs = n
		}
	}
	if v := os.Getenv("JETSON_BACKOFF_MAX_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.BackoffMaxMs = n
		}
	}
	if v := os.Getenv("JETSON_BACKOFF_FACTOR"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.BackoffFactor = f
		}
	}
	if v := os.Getenv("JETSON_BACKOFF_JITTER"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.BackoffJitter = &b
		}
	}
	if v := os.Getenv("JETSON_COOLDOWN_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.NotificationCooldownSeconds = n
		}
	}
	if v := os.Getenv("JETSON_USER_AGENT"); v != "" {
		c.UserAgent = v
	}
}

func (c *ConfigJetson) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.RetryMaxTries < 1 {
		c.RetryMaxTries = 1
	}
	if c.BackoffMinMs < 0 || c.BackoffMaxMs < 0 {
		return errors.New("backoff values must be >= 0")
	}
	if c.BackoffMaxMs < c.BackoffMinMs {
		return fmt.Errorf("backoff_max_ms (%d) < backoff_min_ms (%d)", c.BackoffMaxMs, c.BackoffMinMs)
	}
	if c.BackoffFactor <= 0 {
		return errors.New("backoff_factor must be > 0")
	}
	if len(c.Urls) == 0 {
		return errors.New("urls: at least one target is required")
	}
	for i, u := range c.Urls {
		if u.URL == "" {
			return fmt.Errorf("urls[%d].url is empty", i)
		}
		if u.Scheduler != "" {
			if _, err := cron.ParseStandard(u.Scheduler); err != nil {
				return fmt.Errorf("urls[%d].scheduler invalid: %w", i, err)
			}
		}
	}
	return nil
}

func (c *ConfigJetson) EffectiveTelegram() *TelegramCfg {
	var out TelegramCfg
	if c.Notifications != nil && c.Notifications.Telegram != nil {
		out = *c.Notifications.Telegram
	}

	if out.Token == "" || out.ChatID == 0 {
		return nil
	}
	return &out
}

func (c *ConfigJetson) EffectiveSlack() *SlackCfg {
	var out SlackCfg
	if c.Notifications != nil && c.Notifications.Slack != nil {
		out = *c.Notifications.Slack
	}

	if out.Webhook == "" && out.Token == "" {
		return nil
	}
	return &out
}

func validateFile(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("'%s' is a directory", path)
	}
	return nil
}
