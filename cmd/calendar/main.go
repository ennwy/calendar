package main

import (
	"context"
	"errors"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/logger"
	intergrpc "github.com/ennwy/calendar/internal/server/grpc"
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
	server := intergrpc.NewServer(ctx, l, calendar, addr)

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
