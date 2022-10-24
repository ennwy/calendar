package grpcapi

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"net/http"

	"github.com/ennwy/calendar/internal/app"

	api "github.com/ennwy/calendar/internal/server"
	pb "github.com/ennwy/calendar/internal/server/grpc/google"
	gg "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type GRPCServer struct {
	Ctx        context.Context
	Server     *grpc.Server
	HTTPServer *http.Server
	App        app.Storage
	Addr       string
	pb.UnimplementedStorageServer
}

var _ pb.StorageServer = (*GRPCServer)(nil)

type GRPCLog struct {
	l api.Logger
}

func (l *GRPCLog) Log(data ...interface{}) error {
	l.l.Info(data)
	return nil
}

var l api.Logger

func NewServer(ctx context.Context, log api.Logger, app app.Storage, addr string) *GRPCServer {
	l = log

	s := &GRPCServer{
		Ctx:    ctx,
		App:    app,
		Server: grpc.NewServer(),
		Addr:   addr,
	}

	mux := gg.NewServeMux()

	if err := pb.RegisterStorageHandlerServer(s.Ctx, mux, s); err != nil {
		l.Fatal(err)
	}

	s.HTTPServer = &http.Server{
		Handler: api.LogMiddleware(mux, l),
	}

	pb.RegisterStorageServer(s.Server, s)

	return s
}

func (s *GRPCServer) Start(ctx context.Context) error {
	if err := s.App.Connect(ctx); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	l.Error("[ + ] GRPC STARTED:", err)
	err = s.HTTPServer.Serve(lis)

	return err
}

func (s *GRPCServer) Stop(ctx context.Context) error {
	if err := s.App.Close(ctx); err != nil {
		return err
	}

	s.Server.GracefulStop()
	err := s.HTTPServer.Shutdown(ctx)
	l.Error("\n[ + ] GRPC STOPPED:", err)

	return err
}
