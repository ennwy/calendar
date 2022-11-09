package app

import (
	"context"
	"github.com/ennwy/calendar/internal/storage"
	"time"
)

type Logger interface {
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Fatal(...any)
}

// Database permissions configurations below

type connectCloser interface {
	Connect(context.Context) error
	Close(context.Context) error
}

type Storage interface {
	connectCloser
	CreateEvent(ctx context.Context, event *storage.Event) error
	UpdateEvent(ctx context.Context, event *storage.Event) error
	DeleteEvent(ctx context.Context, eventID int64) error
	ListUserEvents(ctx context.Context, username string) ([]storage.Event, error)
}

type listener interface {
	ListUpcoming(ctx context.Context, until time.Duration) ([]storage.Event, error)
}

type cleaner interface {
	Clean(ctx context.Context, t time.Duration) error
}

type CleanListener interface {
	connectCloser
	cleaner
	listener
}

type App struct {
	l Logger
	Storage
}

func New(logger Logger, storage Storage) *App {
	return &App{
		l:       logger,
		Storage: storage,
	}
}

func (a *App) Connect(ctx context.Context) error {
	return a.Storage.Connect(ctx)
}

func (a *App) Close(ctx context.Context) error {
	return a.Storage.Close(ctx)
}

func (a *App) CreateEvent(ctx context.Context, e *storage.Event) error {
	return a.Storage.CreateEvent(ctx, e)
}

func (a *App) ListUserEvents(ctx context.Context, username string) ([]storage.Event, error) {
	return a.Storage.ListUserEvents(ctx, username)
}

func (a *App) UpdateEvent(ctx context.Context, e *storage.Event) error {
	return a.Storage.UpdateEvent(ctx, e)
}

func (a *App) DeleteEvent(ctx context.Context, eventID int64) error {
	return a.Storage.DeleteEvent(ctx, eventID)
}
