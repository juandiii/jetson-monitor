package logging

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type StandardLogger struct {
	sugar *zap.SugaredLogger
}

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})

	Sync() error
}

func NewLogger() *StandardLogger {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	var lvl zapcore.Level
	switch levelStr {
	case "CRITICAL", "ERROR":
		lvl = zapcore.ErrorLevel
	case "WARNING":
		lvl = zapcore.WarnLevel
	case "NOTICE", "INFO":
		lvl = zapcore.InfoLevel
	case "DEBUG":
		lvl = zapcore.DebugLevel
	default:
		lvl = zapcore.InfoLevel
	}

	format := strings.ToLower(os.Getenv("LOG_FORMAT"))

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, zapcore.Lock(os.Stderr), lvl)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := logger.Sugar()

	return &StandardLogger{sugar: sugar}
}

func (l *StandardLogger) Debugf(format string, args ...interface{}) { l.sugar.Debugf(format, args...) }
func (l *StandardLogger) Infof(format string, args ...interface{})  { l.sugar.Infof(format, args...) }
func (l *StandardLogger) Errorf(format string, args ...interface{}) { l.sugar.Errorf(format, args...) }

func (l *StandardLogger) Debug(args ...interface{}) { l.sugar.Debug(args...) }
func (l *StandardLogger) Info(args ...interface{})  { l.sugar.Info(args...) }
func (l *StandardLogger) Error(args ...interface{}) { l.sugar.Error(args...) }

func (l *StandardLogger) Sync() error { return l.sugar.Sync() }
