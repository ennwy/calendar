package server

import (
	"errors"
	"github.com/ennwy/calendar/internal/app"
)

var (
	ErrTime = errors.New(`"Start" or "Finish" parameter is invalid`)
)

type Logger interface {
	app.Logger
}

type Application interface {
	app.Storage
}
