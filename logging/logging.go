package logging

import (
	"os"

	"github.com/op/go-logging"
)

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05} :: %{level} %{color:reset} %{message}`,
)

var Logger = NewLogger()

type StandardLogger struct {
	*logging.Logger
}

func NewLogger() *StandardLogger {
	var baseLogger = &logging.Logger{}

	var standardLogger = &StandardLogger{baseLogger}
	logging.SetFormatter(format)

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
