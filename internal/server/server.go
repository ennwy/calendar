package server

import (
	"errors"
	"github.com/ennwy/calendar/internal/app"
)

const TimeLayout = "2006-01-02T15:04:05Z07:00"

var (
	ErrTime = errors.New(`"Start" or "Finish" parameter is invalid`)
)

type Logger interface {
	app.Logger
}

type Application interface {
	app.Storage
}
