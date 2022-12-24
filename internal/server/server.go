package server

import (
	"errors"
	"github.com/ennwy/calendar/internal/app"
	"time"
)

const (
	TimeLayout = time.RFC3339Nano

	Day   = time.Hour * 24
	Week  = 7 * Day
	Month = 30 * Day
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

func GetUntil(n int64) time.Duration {
	switch n {
	case 0:
		return Day
	case 1:
		return Week
	case 2:
		return Month
	default:
		return Day
	}
}
