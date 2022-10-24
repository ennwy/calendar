package httpapi

import (
	"bytes"
	"fmt"
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

	TimeLayout = "2006-01-02T15:04:05.999999999Z"
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

	startTime, err := time.ParseInLocation(TimeLayout, e.Start, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("parsing start time: %w", err)
	}

	finishTime, err := time.ParseInLocation(TimeLayout, e.Finish, time.UTC)
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
		b.WriteString(e[i].Start.Format(TimeLayout))
		b.WriteByte(' ')
		b.WriteString(e[i].Finish.Format(TimeLayout))
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

func (s *Server) Create(w http.ResponseWriter, r *http.Request) {
	eventString := &EventString{
		Owner:  r.FormValue(argOwner),
		Title:  r.FormValue(argTitle),
		Start:  r.FormValue(argStart),
		Finish: r.FormValue(argFinish),
	}
	l.Info("adjf;alsdfj")
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
	owner := r.FormValue(argOwner)
	l.Info("adjf;alsdfj")
	l.Info(owner)

	ownerID, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	e, err := s.App.ListEvents(s.Ctx, ownerID)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	printEvents(w, e...)
}

func (s *Server) Update(w http.ResponseWriter, r *http.Request) {
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

	if err = s.App.UpdateEvent(s.Ctx, *event); err != nil {
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
