package main

import (
	"github.com/gofiber/fiber"
	"github.com/juandiii/jetson-monitor/api"
	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/logging"
	"github.com/juandiii/jetson-monitor/scheduler"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func main() {

	log := logging.NewLogger()

	s := &Scheduler{
		cron: cron.New(),
	}

	config := &config.ConfigJetson{
		Logger: log,
	}

	config, err := config.LoadConfig()

	if err != nil {
		panic(err)
	}

	for _, conf := range config.Urls {

		s.cron.AddJob(conf.Scheduler, scheduler.New(conf, config))
	}

	s.cron.Start()

	app := fiber.New(&fiber.Settings{
		DisableStartupMessage: true,
	})

	api.InitializeRoute(app)

	log.Debugf("HTTP start :: listening port: %d", 38080)

	app.Listen(38080)

}
