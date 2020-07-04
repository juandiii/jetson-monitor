package request

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"
)

type Request struct {
	http http.Client
}

func RequestServer(c config.URL, ns []notification.CommandProvider, log *logging.StandardLogger) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

	http := http.Client{
		Timeout: time.Duration(30 * time.Second),
	}

	resp, err := http.Get(c.URL)

	if err != nil {
		log.Errorf("%s \n", c.URL)
		log.Error("Host Unrecheable")

		for _, n := range ns {
			n.SendMessage(&notification.Message{
				Text: fmt.Sprintf("The server %s is down", c.URL),
			})
		}

		return
	}

	if c.StatusCode != nil && *c.StatusCode != resp.StatusCode {
		log.Errorf("%s \n", c.URL)
		return
	}

	log.Debugf("[OK] %s", c.URL)

	return

}
