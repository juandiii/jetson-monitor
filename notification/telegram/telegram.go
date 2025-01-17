package telegram

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"
)

type Telegram struct {
	httpClient    http.Client
	URL           string
	TelegramToken string
	Logger        *logging.StandardLogger
}

func New(c config.URL, log *logging.StandardLogger) notification.CommandProvider {
	return &Telegram{
		httpClient: http.Client{
			Timeout: time.Duration(30 * time.Second),
		},
		URL:           "https://api.telegram.org/bot" + c.TelegramToken,
		TelegramToken: c.TelegramToken,
		Logger:        log,
	}
}

func (t *Telegram) SendMessage(data *notification.Message) error {
	if t.TelegramToken != "" {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(&SendMessage{
			ChatID: 124644812,
			Text:   data.Text,
		})

		req, _ := http.NewRequest("POST", t.URL+"/sendMessage", buf)
		req.Header.Set("Content-Type", "application/json")

		if data != nil {
			t.Logger.Debug("Sending Message to Slack")
			res, e := t.httpClient.Do(req)

			if e != nil {
				return e
			}

			defer res.Body.Close()

		}
	}

	return nil
}
