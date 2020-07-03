package scheduler

import (
	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/notification"
	"github.com/juandiii/jetson-monitor/notification/slack"
	"github.com/juandiii/jetson-monitor/notification/telegram"
	"github.com/juandiii/jetson-monitor/request"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	conf                  config.URL
	NotificationProviders []notification.CommandProvider
}

func New(c config.URL) cron.Job {
	return &Scheduler{
		conf: c,
		NotificationProviders: []notification.CommandProvider{
			slack.New(c),
			telegram.New(c),
		},
	}
}

func (s *Scheduler) Run() {
	request.RequestServer(s.conf, s.NotificationProviders)
}
