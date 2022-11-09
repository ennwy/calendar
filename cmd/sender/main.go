package main

import (
	"context"
	"flag"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	"github.com/ennwy/calendar/internal/notification/sender"
	"os/signal"
	"syscall"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "/etc/calendar/scheduler_config.yml", "Path to configuration file")
}

var l app.Logger

func main() {
	flag.Parse()

	c, err := NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	l = logger.New(c.Logger.Level, c.Logger.OutputPath)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	s, err := sender.NewSender(ctx, l, c.MQ)
	if err != nil {
		l.Error(err)
		return
	}

	go func() {
		<-ctx.Done()

		if err := s.Stop(); err != nil {
			l.Error("Scheduler stop:", err)
		}
		l.Info("[ + ] Scheduler stopped")
	}()

	if err = s.Start(); err != nil {
		l.Error(err)
		return
	}
}
