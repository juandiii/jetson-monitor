package telegram

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"
	"github.com/juandiii/jetson-monitor/sdk"
	"github.com/juandiii/jetson-monitor/version"
)

type Telegram struct {
	Client *sdk.Client
	URL    string
	ChatID int64
	Logger logging.Logger
}

type SendMessageReq struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type TgResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
	ErrorCode   int    `json:"error_code,omitempty"`
}

func New(cfg *config.TelegramCfg, log logging.Logger, client *sdk.Client) notification.CommandProvider {
	if cfg == nil || strings.TrimSpace(cfg.Token) == "" {
		if log != nil {
			log.Debug("telegram: missing token; provider disabled")
		}
		return nil
	}

	chatID, ok := parseChatID(cfg.ChatID)
	if !ok || chatID == 0 {
		if log != nil {
			log.Debug("telegram: empty ChatID; skip provider")
		}
		return nil
	}

	if client == nil {
		httpc := &http.Client{Timeout: 30 * time.Second}
		cli, err := sdk.New(
			httpc,
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

	token := strings.TrimSpace(cfg.Token)
	base := "https://api.telegram.org/bot" + token

	return &Telegram{
		Client: client,
		URL:    base,
		ChatID: chatID,
		Logger: log,
	}
}

func (t *Telegram) SendMessage(m *notification.Message) error {
	if t == nil || t.Client == nil || t.ChatID == 0 || m == nil || m.Text == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := &SendMessageReq{
		ChatID:    t.ChatID,
		Text:      m.Text,
		ParseMode: "",
	}

	req, err := t.Client.NewRequest(ctx, http.MethodPost, t.URL+"/sendMessage", reqBody)
	if err != nil {
		return err
	}

	var respBody TgResponse
	_, err = t.Client.Do(ctx, req, &respBody)
	if err != nil {
		return err
	}
	if !respBody.OK {
		if respBody.Description != "" {
			return fmt.Errorf("telegram: %s (code=%d)", respBody.Description, respBody.ErrorCode)
		}
		return fmt.Errorf("telegram: non-OK response (code=%d)", respBody.ErrorCode)
	}

	if t.Logger != nil {
		t.Logger.Debug("Sent message to Telegram")
	}
	return nil
}

func parseChatID(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, false
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}
