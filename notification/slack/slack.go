package slack

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"
	"github.com/juandiii/jetson-monitor/sdk"
	"github.com/juandiii/jetson-monitor/version"
)

type Slack struct {
	Client *sdk.Client
	URL    string
	Logger logging.Logger
}

func New(cfg *config.SlackCfg, log logging.Logger, client *sdk.Client) notification.CommandProvider {
	if cfg == nil || (cfg.Webhook == "" && cfg.Token == "") {
		if log != nil {
			log.Debug("slack: missing webhook/token; provider disabled")
		}
		return nil
	}

	if client == nil {
		cli, err := sdk.New(
			&http.Client{
				Timeout: 30 * time.Second,
			},
			sdk.SetUserAgent("jetson-monitor server-check/"+version.Version),
			sdk.WithRetryAndBackoffs(sdk.RetryConfig{
				RetryMax:     4,
				RetryWaitMin: sdk.PtrTo(1.0),
				RetryWaitMax: sdk.PtrTo(30.0),
			}),
		)
		if err == nil {
			client = cli
		}
	}

	url := strings.TrimSpace(cfg.Webhook)
	if url == "" {
		url = "https://hooks.slack.com/services/" + strings.TrimSpace(cfg.Token)
	}
	if url == "" {
		if log != nil {
			log.Debug("slack: empty webhook after normalization; provider disabled")
		}
		return nil
	}

	return &Slack{Client: client, URL: url, Logger: log}
}

func (s *Slack) SendMessage(data *notification.Message) error {

	if s == nil || s.Client == nil || data == nil || data.Text == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := s.Client.NewRequest(ctx, http.MethodPost, s.URL, &Message{Text: data.Text})
	if err != nil {
		return err
	}

	_, err = s.Client.Do(ctx, req, io.Discard)
	if err != nil {

		return err
	}

	if s.Logger != nil {
		s.Logger.Debug("Sent message to Slack")
	}
	return nil
}
