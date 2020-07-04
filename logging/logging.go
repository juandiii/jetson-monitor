package logging

import (
	"os"

	"github.com/op/go-logging"
)

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05} :: %{level} %{color:reset} %{message}`,
)

type StandardLogger struct {
	*logging.Logger
}

func NewLogger() *StandardLogger {
	var baseLogger = &logging.Logger{}

	backend := logging.NewLogBackend(os.Stderr, "", 0)

	var standardLogger = &StandardLogger{baseLogger}
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter)
	logging.SetLevel(ParseLevel(os.Getenv("LOG_LEVEL")), "")

	return standardLogger
}

func ParseLevel(level string) logging.Level {
	switch level {
	case "CRITICAL":
		return 0
	case "ERROR":
		return 1
	case "WARNING":
		return 2
	case "NOTICE":
		return 3
	case "INFO":
		return 4
	default:
		return 5
	}
}
