package main

import (
	"context"
	"flag"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	s "github.com/ennwy/calendar/internal/notification/scheduler"
	sqlstorage "github.com/ennwy/calendar/internal/storage/sql"
	"os/signal"
	"syscall"
	"time"
)

const (
	day  = 24 * time.Hour
	year = day * 365
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

	var storage app.CleanListener = sqlstorage.New(l)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	sched, err := s.New(ctx, storage, l, c.MQ)
	if err != nil {
		l.Fatal("nil scheduler:", err)
	}

	go func() {
		<-ctx.Done()
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
		l.Error(err)
	}
}
