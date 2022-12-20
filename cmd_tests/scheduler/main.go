package main

import (
	"context"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	s "github.com/ennwy/calendar/internal/notification/scheduler"
	sqlstorage "github.com/ennwy/calendar/internal/storage/sql"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var l app.Logger

func main() {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}

	l = logger.New(config.Logger.Level, config.Logger.OutputPath)
	l.Error("scheduler: config:", err)
	l.Info("[ + ] CONFIG:", config)

	var storage app.CleanListener = sqlstorage.New(l)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	sched, err := s.New(ctx, storage, l, config.MQ)
	if err != nil {
		l.Fatal("nil scheduler:", err)
	}

	go func() {
		<-ctx.Done()
		l.Info("[ + ] stop: ctx canceled")
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		if err := storage.Close(ctx); err != nil {
			l.Error("close db conn: ", err)
		}
		if err := sched.Stop(); err != nil {
			l.Error("Scheduler stop:", err)
		}
		l.Info("[ + ] Scheduler stopped")
	}()

	if err := sched.Start(); err != nil {
		l.Error("scheduler: start:", err)
		//cancel()
		os.Exit(1) //nolint:gocritic
	}
}