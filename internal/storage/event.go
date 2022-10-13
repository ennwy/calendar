package storage

import (
	"time"
)

type Event struct {
	Start   time.Time
	Finish  time.Time
	Title   string
	OwnerID int64
	ID      int64
}
