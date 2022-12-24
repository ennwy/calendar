package main

import (
	"context"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	"github.com/ennwy/calendar/internal/notification/sender"
	"os"
	"os/signal"
	"syscall"
)

var l app.Logger

func main() {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}

	l = logger.New(config.Logger.Level, config.Logger.OutputPath)

	l.Error("sender: config:", err)
	l.Info("[ + ] CONFIG:", config)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	s, err := sender.NewSender(ctx, l, config.MQ)
	l.Info("NewSender: err:", err, "\nSender:", s)

	if err != nil {
		l.Error(err)
		os.Exit(1) //nolint:gocritic
	}

	go func() {
		<-ctx.Done()
		l.Info("[ + ] stop: ctx canceled")

		if err := s.Stop(); err != nil {
			l.Error("Sender stop:", err)
		}
		l.Info("[ + ] Sender stopped")
	}()

	messageCh, err := s.Start()
	if err != nil {
		l.Error("sender: start", err)
		//cancel()
		os.Exit(1) //nolint:gocritic
	}
	s.PrintMessages(messageCh)

	<-ctx.Done()
}
