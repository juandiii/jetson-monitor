package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/notification"
)

type Slack struct {
	httpClient http.Client
	URL        string
	SlackToken string
}

func New(c config.URL) notification.CommandProvider {
	return &Slack{
		httpClient: http.Client{
			Timeout: time.Duration(30 * time.Second),
		},
		URL:        "https://hooks.slack.com/services/" + c.SlackToken,
		SlackToken: c.SlackToken,
	}
}

func (s *Slack) SendMessage(data *notification.Message) error {

	if s.SlackToken != "" {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(&Message{
			Text: data.Text,
		})

		req, _ := http.NewRequest("POST", s.URL, buf)
		req.Header.Set("Content-Type", "application/json")

		if data != nil {
			res, e := s.httpClient.Do(req)

			if e != nil {
				return e
			}

			defer res.Body.Close()

			fmt.Println("response status: ", res.Status)
		}
	}

	return nil
}
