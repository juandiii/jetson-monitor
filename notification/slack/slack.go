package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"
)

type Slack struct {
	httpClient http.Client
	URL        string
	SlackToken string
	Logger     *logging.StandardLogger
}

func New(c config.URL, log *logging.StandardLogger) notification.CommandProvider {
	return &Slack{
		httpClient: http.Client{
			Timeout: time.Duration(30 * time.Second),
		},
		URL:        "https://hooks.slack.com/services/" + c.SlackToken,
		SlackToken: c.SlackToken,
		Logger:     log,
	}
}

func (s *Slack) SendMessage(data *notification.Message) error {

	log := s.Logger

	if s.SlackToken != "" {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(&Message{
			Text: data.Text,
		})

		req, _ := http.NewRequest("POST", s.URL, buf)
		req.Header.Set("Content-Type", "application/json")

		if data != nil {
			log.Debug("Sending Message to Slack")
			res, e := s.httpClient.Do(req)

			if e != nil {
				return e
			}

			defer res.Body.Close()

		}
	}

	return nil
}
