package server

import (
	"context"
	"google.golang.org/grpc/reflection"
	"log"
	"net"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	"google.golang.org/grpc"
)

type Logger interface {
	app.Logger
}

type Application interface {
	app.Storage
}

type Server struct {
	server     *grpc.Server
	app        Application
	host, port string
}

var l Logger

func NewServer(logger Logger, app Application, host, port string) *Server {
	l = logger

	return &Server{
		app:    app,
		server: grpc.NewServer(),
		host:   host,
		port:   port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.app.Connect(ctx); err != nil {
		return err
	}

	_, err := net.Listen("tcp", net.JoinHostPort(s.host, s.port))
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	reflection.Register(s.server)

	l.Info("[ + ] GRPC STARTED")
	return err
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.app.Close(ctx); err != nil {
		return err
	}

	l.Info("\n[ + ] GRPC STOPPED")
	s.server.GracefulStop()
	return nil
}
