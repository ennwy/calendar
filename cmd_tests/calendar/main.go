package main

import (
	"context"
	"errors"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	intergrpc "github.com/ennwy/calendar/internal/server/grpc"
	interhttp "github.com/ennwy/calendar/internal/server/http"
	//pb "github.com/ennwy/calendar/internal/server/grpc/google"
	s "github.com/ennwy/calendar/internal/storage/sql"
	"time"

	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var l app.Logger

func main() {
	config := NewConfig()
	l = logger.New(config.Logger.Level, config.Logger.OutputPath)

	l.Info("[ + ] CONFIG:", config)
	calendar := app.New(l, s.New(l))

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	addr := net.JoinHostPort(config.HTTP.Host, config.HTTP.Port)
	var serverHTTP app.Server = interhttp.NewServer(ctx, l, calendar, addr)

	addr = net.JoinHostPort(config.GRPC.Host, config.GRPC.Port)
	var serverGRPC app.Server = intergrpc.NewServer(ctx, l, calendar, addr)

	go func() {
		<-ctx.Done()
		l.Info("[ + ] stop: ctx canceled")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		go func() {
			if err := serverGRPC.Stop(ctx); err != nil {
				l.Error("failed to stop grpc server:", err)
			}
		}()

		if err := serverHTTP.Stop(ctx); err != nil {
			l.Error("failed to stop http server:", err)
		}

		l.Info("[ + ] Calendar stopped")
	}()

	l.Info("[ + ] calendar is running...")

	go func() {
		err := serverGRPC.Start(ctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("grpc server error: ", err)
			//cancel()
			os.Exit(1) //nolint:gocritic
		}
	}()

	err := serverHTTP.Start(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Error("http server error: ", err)
		//cancel()
		os.Exit(1) //nolint:gocritic
	}

}
