package server

import (
	"github.com/ennwy/calendar/internal/app"
)

type Logger interface {
	app.Logger
}

type Application interface {
	app.Storage
}
