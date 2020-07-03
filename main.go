package main

import (
	"github.com/gofiber/fiber"
	"github.com/juandiii/jetson-monitor/api"
	"github.com/juandiii/jetson-monitor/config"
	"github.com/juandiii/jetson-monitor/scheduler"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func main() {

	s := &Scheduler{
		cron: cron.New(),
	}

	config := &config.ConfigAmanda{}

	config, err := config.LoadConfig()

	if err != nil {
		panic(err)
	}

	for _, conf := range config.Urls {

		s.cron.AddJob(conf.Scheduler, scheduler.New(conf))
	}

	s.cron.Start()

	app := fiber.New()

	api.InitializeRoute(app)

	app.Listen(38080)

}
