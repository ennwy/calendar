package httpapi

import (
	"bytes"
	"fmt"
	api "github.com/ennwy/calendar/internal/server"
	"github.com/ennwy/calendar/internal/storage"
	"net/http"
	"strconv"
	"time"
)

const (
	argID     = "id"
	argOwner  = "owner"
	argTitle  = "title"
	argStart  = "start"
	argFinish = "finish"
	argNotify = "notify"

	TimeLayout = "2006-01-02T15:04:05.999999999Z"
)

type EventString struct {
	OwnerName string
	Title     string
	Start     string
	Finish    string
	Notify    string
}

func (e *EventString) convertToEvent() (*storage.Event, error) {
	startTime, err := time.ParseInLocation(TimeLayout, e.Start, time.UTC)

	if err != nil {
		return nil, fmt.Errorf("parsing start time: %w", err)
	}

	finishTime, err := time.ParseInLocation(TimeLayout, e.Finish, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("parsing finish time: %w", err)
	}

	if !startTime.Before(finishTime) {
		return nil, api.ErrTime
	}

	return &storage.Event{
			Start:  startTime,
			Finish: finishTime,
			Title:  e.Title,
			Owner:  storage.User{Name: e.OwnerName},
		},
		nil
}

func printEvents(w http.ResponseWriter, events ...storage.Event) {
	b := bytes.Buffer{}

	for _, e := range events {

		b.WriteString(strconv.FormatInt(e.ID, 10))
		b.WriteByte(' ')
		b.WriteString(e.Owner.Name)
		b.WriteByte(' ')
		b.WriteString(e.Start.Format(TimeLayout))
		b.WriteByte(' ')
		b.WriteString(e.Finish.Format(TimeLayout))
		b.Write([]byte{' ', '"'})
		b.WriteString(e.Title)
		b.Write([]byte{'"', '\n'})

		_, _ = w.Write(b.Bytes())
		b.Reset()
	}
}

func respondAndLog(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	l.Error(err)
}

func (s *Server) Create(w http.ResponseWriter, r *http.Request) {
	eventString := &EventString{
		OwnerName: r.FormValue(argOwner),
		Title:     r.FormValue(argTitle),
		Start:     r.FormValue(argStart),
		Finish:    r.FormValue(argFinish),
		Notify:    r.FormValue(argNotify),
	}
	l.Info(eventString)

	event, err := eventString.convertToEvent()
	if err != nil {
		respondAndLog(w, err)
		return
	}

	if err = s.App.CreateEvent(s.Ctx, event); err != nil {
		respondAndLog(w, err)
	}

	printEvents(w, *event)
}

func (s *Server) List(w http.ResponseWriter, r *http.Request) {
	ownerName := r.FormValue(argOwner)

	e, err := s.App.ListUserEvents(s.Ctx, ownerName)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	printEvents(w, e...)
}

func (s *Server) Update(w http.ResponseWriter, r *http.Request) {
	eventString := &EventString{
		OwnerName: r.FormValue(argOwner),
		Title:     r.FormValue(argTitle),
		Start:     r.FormValue(argStart),
		Finish:    r.FormValue(argFinish),
	}

	event, err := eventString.convertToEvent()
	if err != nil {
		respondAndLog(w, err)
		return
	}

	id, err := strconv.ParseInt(r.FormValue(argID), 10, 64)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	event.ID = id

	if err = s.App.UpdateEvent(s.Ctx, event); err != nil {
		respondAndLog(w, err)
	}
}

func (s *Server) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue(argID), 10, 64)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	if err = s.App.DeleteEvent(s.Ctx, id); err != nil {
		respondAndLog(w, err)
	}
}
