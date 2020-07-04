package scheduler

import (
	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/notification"
	"github.com/juandiii/jetson-monitor/notification/slack"
	"github.com/juandiii/jetson-monitor/request"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	conf                  config.URL
	NotificationProviders []notification.CommandProvider
	Logger                *logging.StandardLogger
}

func New(c config.URL, conf *config.ConfigJetson) cron.Job {
	return &Scheduler{
		conf: c,
		NotificationProviders: []notification.CommandProvider{
			slack.New(c, conf.Logger),
			// telegram.New(c),
		},
		Logger: conf.Logger,
	}
}

func (s *Scheduler) Run() {
	request.RequestServer(s.conf, s.NotificationProviders, s.Logger)
}
