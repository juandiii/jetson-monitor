package request

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jpillora/backoff"
	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"
)

type Request struct {
	http http.Client
}

func RequestServer(c config.URL, ns []notification.CommandProvider, log *logging.StandardLogger) (string, error) {
	b := &backoff.Backoff{
		Jitter: true,
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

	client := &http.Client{}

	req, err := http.NewRequest("GET", c.URL, nil)

	if err != nil {
		log.Error(err)
		return "", errors.New("Received an invalid status code: 500 The service might be experiencing issues")
	}

	for tries := 0; tries < 3; tries++ {
		resp, err := client.Do(req)

		if err != nil {
			d := b.Duration()
			time.Sleep(d)

			if tries == 2 {
				for _, n := range ns {
					n.SendMessage(&notification.Message{
						Text: fmt.Sprintf("The server %s is down", c.URL),
					})
				}

				return "", fmt.Errorf("The server %s is down", c.URL)
			}

			continue

		}

		defer resp.Body.Close()

		if c.StatusCode != nil && *c.StatusCode != resp.StatusCode {
			log.Errorf("%s \n", c.URL)
			return "", errors.New("Received an invalid status code: " + strconv.Itoa(resp.StatusCode) + " The service might be experiencing issues")
		}

		log.Debugf("[OK] %s", c.URL)

		return fmt.Sprintf("[OK] %s", c.URL), nil
	}

	return "", errors.New("The request failed because it wasn't able to reach the service")

}
