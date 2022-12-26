package main

import (
	"context"
	"errors"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	intergrpc "github.com/ennwy/calendar/internal/server/grpc"
	interhttp "github.com/ennwy/calendar/internal/server/http"
	//pb "github.com/ennwy/calendar/internal/server/grpc/google"
	storage "github.com/ennwy/calendar/internal/storage/sql"
	"time"

	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var l app.Logger

func main() {
	if len(os.Args) > 0 && os.Args[0] == "version" {
		printVersion()
		return
	}

	config := NewConfig()
	l = logger.New(config.Logger.Level, config.Logger.OutputPath)

	l.Info("[ + ] CONFIG:", config)
	calendar := app.New(l, storage.New(l))

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var server app.Server
	switch config.Server {
	case "grpc":
		addr := net.JoinHostPort(config.GRPC.Host, config.GRPC.Port)
		server = intergrpc.NewServer(ctx, l, calendar, addr)
	case "http":
		addr := net.JoinHostPort(config.HTTP.Host, config.HTTP.Port)
		server = interhttp.NewServer(ctx, l, calendar, addr)
	default:
		l.Error("SERVER_TYPE var is not correct. http or grpc expected")
		return
	}

	go func() {
		<-ctx.Done()
		l.Info("[ + ] stop: ctx canceled")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			l.Error("failed to stop http server: ", err)
		}

		l.Info("[ + ] Calendar stopped")
	}()

	l.Info("[ + ] calendar is running...")

	err := server.Start(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Error("server error: ", err)
		//cancel()
		os.Exit(1) //nolint:gocritic
	}
}
