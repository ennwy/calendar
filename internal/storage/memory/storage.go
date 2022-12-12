package memstorage

import (
	"context"
	"errors"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/storage"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	dayRounded    = "2006-01-02"
	minuteRounded = "2006-01-02 15:04:05"
)

// Test storage

//type Users map[int64]map[int64]storage.Event

//type Storage struct {
//	mu    sync.RWMutex
//	m     Users
//	maxID int64
//}

type Storage struct {
	m     sync.Map
	mu    *sync.RWMutex
	maxID int64
}

var _ app.Storage = (*Storage)(nil)
var _ app.CleanListener = (*Storage)(nil)

func New() *Storage {
	return &Storage{
		m:  sync.Map{},
		mu: &sync.RWMutex{},
	}
}

func (s *Storage) Connect(_ context.Context) error {
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	return nil
}

func (s *Storage) CreateEvent(_ context.Context, e *storage.Event) error {
	atomic.AddInt64(&s.maxID, 1)
	atomic.StoreInt64(&e.ID, atomic.LoadInt64(&s.maxID))
	s.m.Store(e.ID, *e)
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, e *storage.Event) error {
	id := atomic.LoadInt64(&e.ID)

	v, ok := s.m.Load(id)
	if !ok {
		return errors.New("event was not exist before")
	}
	event, ok := v.(storage.Event)
	if !ok {

		return errors.New("cast to event")
	}

	//StoreString(&event.Owner.Name, e.Owner.Name)
	s.mu.Lock()
	event.Owner.Name = e.Owner.Name
	log.Println("event name test:", event.Owner.Name)
	s.mu.Unlock()

	s.mu.RLock()
	eValue := *e
	s.mu.RUnlock()

	s.m.Store(id, eValue)

	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, eventID int64) error {
	s.m.Delete(eventID)

	return nil
}

func (s *Storage) ListUserEvents(_ context.Context, username string) ([]storage.Event, error) {
	events := make([]storage.Event, 0, 1)

	s.m.Range(func(key, value any) bool {
		v, ok := value.(storage.Event)
		if !ok {
			return true
		}

		s.mu.RLock()
		name := v.Owner.Name
		s.mu.RUnlock()

		if username == name {
			events = append(events, v)
		}

		return true
	})

	return events, nil
}

func (s *Storage) ListUpcoming(_ context.Context, until time.Duration) ([]storage.Event, error) {
	if until < 0 {
		until = -until
	}

	current := time.Now().Round(time.Minute)
	bound := current.Add(until)
	currentU := current.Unix()
	boundU := bound.Unix()

	events := make([]storage.Event, 0, 1)

	s.m.Range(func(key, value any) bool {
		v, ok := value.(storage.Event)
		if !ok {
			return true
		}

		notify := v.Start.Add(time.Duration(v.Notify)).Unix()

		if currentU <= notify && notify > boundU {
			events = append(events, v)
		}

		return true
	})

	return events, nil
}

func (s *Storage) Clean(_ context.Context, ago time.Duration) error {
	if ago > 0 {
		ago = -ago
	}

	finishedAgo := time.Now().Round(24 * time.Hour).UTC().Add(ago)
	s.m.Range(func(key, value any) bool {
		v, ok := value.(storage.Event)
		if !ok {
			return true
		}

		if finishedAgo.Before(v.Finish) {
			s.m.Delete(key)
		}
		return true
	})

	return nil
}
