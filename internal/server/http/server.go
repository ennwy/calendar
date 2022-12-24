package httpapi

import (
	"context"
	api "github.com/ennwy/calendar/internal/server"
	"github.com/julienschmidt/httprouter"
	"net/http"

	"github.com/ennwy/calendar/internal/app"
)

const (
	CreateEventPath = "/create"
	UpdateEventPath = "/update"
	DeleteEventPath = "/delete"
	ListEventsPath  = "/list"
)

type Server struct {
	Ctx context.Context
	S   *http.Server
	App app.Storage
}

var l api.Logger

func NewServer(ctx context.Context, log api.Logger, App api.Application, addr string) *Server {
	l = log

	l.Info("http server addr:", addr)
	s := &Server{
		Ctx: ctx,
		App: App,
		S: &http.Server{
			Addr:    addr,
			Handler: nil,
		},
	}

	r := httprouter.New()
	r.HandlerFunc(http.MethodGet, CreateEventPath, s.Create)
	r.HandlerFunc(http.MethodGet, UpdateEventPath, s.Update)
	r.HandlerFunc(http.MethodGet, DeleteEventPath, s.Delete)
	r.HandlerFunc(http.MethodGet, ListEventsPath, s.List)

	handler := api.LogMiddleware(r, l)
	s.S.Handler = handler

	return s
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.App.Connect(ctx); err != nil {
		return err
	}
	l.Info("[ + ] Successfully connected to db")
	l.Info("[ + ] HTTP STARTED")
	return s.S.ListenAndServe()

}

func (s *Server) Stop(ctx context.Context) (err error) {
	if err = s.App.Close(ctx); err != nil {
		return err
	}

	l.Info("\n[ + ] HTTP STOPPED")

	return s.S.Shutdown(ctx)
}
