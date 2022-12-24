package main

import (
	"context"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	"github.com/ennwy/calendar/internal/notification/scheduler"
	"github.com/ennwy/calendar/internal/notification/sender"
	"github.com/streadway/amqp"
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

	l.Error("s: config:", err)
	l.Info("[ + ] CONFIG:", config)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	receiver, err := sender.NewSender(ctx, l, config.MQConsume)
	l.Info("New receiver: err:", err, "\nReceiver:", receiver)
	if err != nil {
		l.Error(err)
		os.Exit(1) //nolint:gocritic
	}

	s, err := scheduler.NewQueue(config.MQProduce)
	l.Info("New s: err:", err, "\nSender:", s)
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}

	go func() {
		<-ctx.Done()
		l.Info("[ + ] stop: ctx canceled")

		if err := receiver.Stop(); err != nil {
			l.Error("Receiver stop:", err)
		}
		if err := s.Close(); err != nil {
			l.Error("Sender stop:", err)
		}
		l.Info("[ + ] Sender stopped")
	}()

	messageCh, err := receiver.Start()
	if err != nil {
		l.Error("s: start", err)
		//cancel()
		os.Exit(1) //nolint:gocritic
	}

	for message := range messageCh {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err = s.Ch.Publish(
			"",
			s.Q.Name,
			s.Opts.Mandatory,
			s.Opts.Immediate,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        message.Body,
			},
		)

		l.Info("Published received event:")
		if err != nil {
			l.Error("ERR publishing received event:", err)
		}
		l.Info(message.Body, "\n")

		if err := message.Ack(false); err != nil {
			l.Error("message ack:", err)
		}
	}

	<-ctx.Done()
}
