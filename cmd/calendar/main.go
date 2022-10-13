package main

import (
	"context"
	"errors"
	"flag"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	interhttp "github.com/ennwy/calendar/internal/server/http"
	memstorage "github.com/ennwy/calendar/internal/storage/memory"
	sqlstorage "github.com/ennwy/calendar/internal/storage/sql"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "/etc/calendar/config.yml", "Path to configuration file")
}

var l app.Logger

func main() {
	flag.Parse()
	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	l = logger.New(config.Logger.Level, config.Logger.OutputPath)

	var storage app.Storage

	switch config.Storage {
	case "sql":
		storage = sqlstorage.New(l)
	case "mem", "memory":
		storage = memstorage.New()
	default:
		l.Fatal("invalid storage type was given")
	}

	calendar := app.New(l, storage)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	server := interhttp.NewServer(ctx, l, calendar, config.HTTP.Host, config.HTTP.Port)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			l.Error("failed to stop http server: ", err)
		}
	}()

	l.Info("[ + ] calendar is running...")

	err = server.Start(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Error("server error: ", err)
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
