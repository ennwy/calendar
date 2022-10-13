package app

import (
	"context"

	"github.com/ennwy/calendar/internal/storage"
)

type Logger interface {
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Fatal(...any)
}

type Storage interface {
	Connect(context.Context) error
	Close(context.Context) error
	CreateEvent(context.Context, *storage.Event) error
	ListEvents(context.Context, int64) ([]storage.Event, error)
	UpdateEvent(context.Context, storage.Event) error
	DeleteEvent(context.Context, int64) error
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

func (a *App) ListEvents(ctx context.Context, ownerID int64) ([]storage.Event, error) {
	return a.Storage.ListEvents(ctx, ownerID)
}

func (a *App) UpdateEvent(ctx context.Context, e storage.Event) error {
	return a.Storage.UpdateEvent(ctx, e)
}

func (a *App) DeleteEvent(ctx context.Context, eventID int64) error {
	return a.Storage.DeleteEvent(ctx, eventID)
}
