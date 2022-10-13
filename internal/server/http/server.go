package server

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"time"

	"github.com/ennwy/calendar/internal/app"
)

type httpInfo struct {
	date        time.Time
	latency     time.Duration
	ip          string
	httpMethod  string
	path        string
	httpVersion string
	respStatus  string
	userAgent   string
}

func NewHTTPInfo(r *http.Request) *httpInfo {
	latency := time.Now()
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	info := &httpInfo{
		ip:          ip,
		date:        time.Now(),
		httpVersion: r.Proto,
		httpMethod:  r.Method,
		path:        r.URL.Path,
		respStatus:  "mb 200", // r.Response.Status,
		userAgent:   r.UserAgent(),
	}

	if r.Response != nil {
		info.respStatus = r.Response.Status
	}

	info.latency = time.Since(latency)

	return info
}

func (i *httpInfo) String() string {
	return fmt.Sprintf("%s [%s] %s %s %s %q %s %s",
		i.ip,
		i.date.Format("02-01-2006 15:04:05"),
		i.httpMethod,
		i.path,
		i.httpVersion,
		i.respStatus,
		i.latency.String(),
		i.userAgent,
	)
}

type Logger interface {
	app.Logger
}

type Application interface {
	app.Storage
}

type Server struct {
	server *http.Server
	app    Application
	ctx    context.Context
}

var l Logger

func NewServer(ctx context.Context, logger Logger, app Application, host, port string) *Server {
	l = logger

	s := &Server{
		ctx: ctx,
		app: app,
		server: &http.Server{
			Addr:    net.JoinHostPort(host, port),
			Handler: nil,
		},
	}

	r := httprouter.New()
	r.HandlerFunc(http.MethodGet, CreateEventPath, s.create)
	r.HandlerFunc(http.MethodGet, UpdateEventPath, s.update)
	r.HandlerFunc(http.MethodGet, DeleteEventPath, s.delete)
	r.HandlerFunc(http.MethodGet, ListEventsPath, s.list)

	handler := logMiddleware(r)
	s.server.Handler = handler

	return s
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.app.Connect(ctx); err != nil {
		return err
	}
	l.Info("[ + ] Successfully connected to db")
	l.Info("[ + ] HTTP STARTED")
	return s.server.ListenAndServe()

}

func (s *Server) Stop(ctx context.Context) (err error) {
	if err = s.app.Close(ctx); err != nil {
		return err
	}

	l.Info("\n[ + ] HTTP STOPPED")

	return s.server.Shutdown(ctx)
}

func (s *Server) helloWorld(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("some fucking long text to test that fucking system\n"))
}
