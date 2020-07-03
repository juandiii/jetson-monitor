package request

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/notification"
)

const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

type Request struct {
	http http.Client
}

func RequestServer(c config.URL, ns []notification.CommandProvider) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

	http := http.Client{
		Timeout: time.Duration(30 * time.Second),
	}

	resp, err := http.Get(c.URL)

	tmpString := ""

	if err != nil {
		tmpString = "[ERROR] " + c.URL + "\n"
		fmt.Printf(ErrorColor, tmpString)
		fmt.Println("Host Unrecheable")

		for _, n := range ns {
			n.SendMessage(&notification.Message{
				Text: fmt.Sprintf("The server %s is down", c.URL),
			})
		}

		return
	}

	if c.StatusCode != nil && *c.StatusCode != resp.StatusCode {
		tmpString = "[ERROR] " + c.URL + "\n"
		fmt.Printf(ErrorColor, tmpString)
		return
	}

	tmpString = "[OK] " + c.URL + "\n"
	fmt.Printf(NoticeColor, tmpString)

	return

}
