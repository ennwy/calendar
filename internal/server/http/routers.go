package httpapi

import (
	"encoding/json"
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
	argUntil  = "until"
)

type EventString struct {
	OwnerName string
	Title     string
	Start     string
	Finish    string
	Notify    string
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
}

func (s *Server) List(w http.ResponseWriter, r *http.Request) {
	ownerName := r.FormValue(argOwner)
	untilS := r.FormValue(argUntil)
	var events *storage.Events
	var err error

	if untilS == "" {
		events, err = s.App.ListUserEvents(s.Ctx, ownerName)
	} else {
		n, _ := strconv.ParseInt(untilS, 10, 32)
		until := api.GetUntil(n)
		events, err = s.App.ListUsersUpcoming(s.Ctx, ownerName, until)
	}

	if err != nil {
		respondAndLog(w, err)
	}

	printEvents(w, events)
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
func (e *EventString) convertToEvent() (*storage.Event, error) {
	startTime, err := time.Parse(api.TimeLayout, e.Start)
	if err != nil {
		return nil, fmt.Errorf("parsing start time: %w", err)
	}

	finishTime, err := time.Parse(api.TimeLayout, e.Finish)
	if err != nil {
		return nil, fmt.Errorf("parsing finish time: %w", err)
	}

	if !startTime.Before(finishTime) {
		return nil, api.ErrTime
	}

	notify, err := strconv.ParseInt(e.Notify, 10, 32)

	return &storage.Event{
		Start:  startTime.UTC(),
		Finish: finishTime.UTC(),
		Title:  e.Title,
		Owner:  storage.User{Name: e.OwnerName},
		Notify: int32(notify),
	}, nil
}

func printEvents(w http.ResponseWriter, events *storage.Events) {
	w.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(events)
	if err != nil {
		respondAndLog(w, err)
		return
	}

	if _, err = w.Write(b); err != nil {
		l.Error("http router: printEvents:", err)
	}
}

func respondAndLog(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	l.Error(err)
}
