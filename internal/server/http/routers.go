package server

import (
	"bytes"
	"fmt"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"net/http"
	"strconv"
	"time"
)

const (
	CreateEventPath = "/create"
	UpdateEventPath = "/update"
	DeleteEventPath = "/delete"
	ListEventsPath  = "/list"

	// Query arguments.
	argID     = "id"
	argOwner  = "owner"
	argTitle  = "title"
	argStart  = "start"
	argFinish = "finish"

	timeLayout = "2006-01-02_15:04:05_0000_UTC"
)

type EventString struct {
	Owner  string
	Title  string
	Start  string
	Finish string
}

func (e *EventString) convertToEvent() (*storage.Event, error) {
	ownerID, err := strconv.ParseInt(e.Owner, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing owner id: %w", err)
	}

	startTime, err := time.ParseInLocation(timeLayout, e.Start, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("parsing start time: %w", err)
	}

	finishTime, err := time.ParseInLocation(timeLayout, e.Finish, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("parsing finish time: %w", err)
	}

	return &storage.Event{
			Start:   startTime,
			Finish:  finishTime,
			Title:   e.Title,
			OwnerID: ownerID,
		},
		nil
}

func printEvents(w http.ResponseWriter, e ...storage.Event) {
	b := bytes.Buffer{}

	for i := range e {
		b.WriteString(strconv.FormatInt(e[i].ID, 10))
		b.WriteByte(' ')
		b.WriteString(strconv.FormatInt(e[i].OwnerID, 10))
		b.WriteByte(' ')
		b.WriteString(e[i].Start.Format(timeLayout))
		b.WriteByte(' ')
		b.WriteString(e[i].Finish.Format(timeLayout))
		b.Write([]byte{' ', '"'})
		b.WriteString(e[i].Title)
		b.Write([]byte{'"', '\n'})

		_, _ = w.Write(b.Bytes())
		b.Reset()
	}
}

func respondAndLog(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	l.Error(err)
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	eventString := &EventString{
		Owner:  r.FormValue(argOwner),
		Title:  r.FormValue(argTitle),
		Start:  r.FormValue(argStart),
		Finish: r.FormValue(argFinish),
	}

	event, err := eventString.convertToEvent()
	if err != nil {
		respondAndLog(w, err)
		return
	}

	if err = s.app.CreateEvent(s.ctx, event); err != nil {
		respondAndLog(w, err)
	}

	printEvents(w, *event)
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	owner := r.FormValue(argOwner)
	ownerID, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	e, err := s.app.ListEvents(s.ctx, ownerID)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	printEvents(w, e...)
}

func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	eventString := &EventString{
		Owner:  r.FormValue(argOwner),
		Title:  r.FormValue(argTitle),
		Start:  r.FormValue(argStart),
		Finish: r.FormValue(argFinish),
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

	if err = s.app.UpdateEvent(s.ctx, *event); err != nil {
		respondAndLog(w, err)
	}
}

func (s *Server) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue(argID), 10, 64)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	if err = s.app.DeleteEvent(s.ctx, id); err != nil {
		respondAndLog(w, err)
	}
}
