package httpapi

import (
	"github.com/ennwy/calendar/internal/storage"
	"strconv"
	"time"
)

// PROTOBUF converts INT64 to STRING in json, so we must have same output

type UserString struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

type EventString struct {
	Start  time.Time  `json:"Start"`
	Finish time.Time  `json:"Finish"`
	Owner  UserString `json:"Owner"`
	Title  string     `json:"Title"`
	ID     string     `json:"ID"`
	Notify int32      `json:"Notify"`
}

func ToEventString(e storage.Event) EventString {
	return EventString{
		ID: strconv.FormatInt(e.ID, 10),
		Owner: UserString{
			ID:   strconv.FormatInt(e.Owner.ID, 10),
			Name: e.Owner.Name,
		},
		Title:  e.Title,
		Start:  e.Start,
		Finish: e.Finish,
		Notify: e.Notify, // INT32 is still int32 in protobuf json
	}
}

type EventsString struct {
	Events []EventString
}

func ToEvents(events *storage.Events) *EventsString {
	es := &EventsString{
		Events: make([]EventString, 0, len(events.Events)),
	}

	for _, e := range events.Events {
		es.Events = append(es.Events, ToEventString(e))
	}

	return es
}
