package main

import (
	"context"
	"errors"
	"flag"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	sqlstorage "github.com/ennwy/calendar/internal/storage/sql"
	"time"

	intergrpc "github.com/ennwy/calendar/internal/server/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "/etc/calendar/calendar_config.yaml", "Path to configuration file")
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

	var storage app.Storage = sqlstorage.New(l)

	calendar := app.New(l, storage)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	addr := net.JoinHostPort(config.HTTP.Host, config.HTTP.Port)
	server := intergrpc.NewServer(ctx, l, calendar, addr)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
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
