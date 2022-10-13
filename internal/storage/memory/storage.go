package memstorage

import (
	"context"
	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/storage"
	"sync"
	"sync/atomic"
)

type Users map[int64]map[int64]storage.Event

type Storage struct {
	mu    sync.RWMutex
	m     Users
	maxID int64
}

var _ app.Storage = (*Storage)(nil)

func New() *Storage {
	return &Storage{
		m:     make(Users),
		maxID: 0,
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

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.m[e.OwnerID] == nil {
		s.m[e.OwnerID] = make(map[int64]storage.Event, 1)
	}
	s.m[e.OwnerID][e.ID] = *e

	return nil
}

func (s *Storage) ListEvents(_ context.Context, ownerID int64) ([]storage.Event, error) {
	eventList := make([]storage.Event, 0, len(s.m))

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, event := range s.m[ownerID] {
		eventList = append(eventList, event)
	}

	return eventList, nil
}

func (s *Storage) UpdateEvent(_ context.Context, e storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.m[e.OwnerID] == nil {
		s.m[e.OwnerID] = make(map[int64]storage.Event)
	}
	s.m[e.OwnerID][e.ID] = e
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, eventID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range s.m {
		if _, ok := m[eventID]; ok {
			delete(m, eventID)
		}
	}
	return nil
}
