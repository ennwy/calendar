package integration_tests

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const TimeLayout = "2006-01-02T15:04:05Z07:00"

var addr = "http://" + net.JoinHostPort(os.Getenv("TEST_HOST"), os.Getenv("TEST_PORT"))

//var query = addr + "/%s/%s/%s/%q/%q/%d"

// JSON convert int64 to string, so we need to parse it to int64

type Events map[int64]Event

type UserString struct {
	Name string `json:"Name"`
	ID   string `json:"ID"`
}

type EventString struct {
	Start  time.Time  `json:"Start"`
	Finish time.Time  `json:"Finish"`
	Owner  UserString `json:"Owner"`
	Title  string     `json:"Title"`
	ID     string     `json:"ID"`
	Notify int32      `json:"Notify"`
}

func (e *EventString) ToEvent() Event {
	// WE CAN'T GET ERROR HERE BECAUSE SERVER RETURNS ONLY INT VALUES
	ownerID, _ := strconv.ParseInt(e.Owner.ID, 10, 64)
	eventID, _ := strconv.ParseInt(e.ID, 10, 64)

	return Event{
		ID: eventID,
		Owner: User{
			ID:   ownerID,
			Name: e.Owner.Name,
		},
		Title:  e.Title,
		Start:  e.Start,
		Finish: e.Finish,
		Notify: e.Notify,
	}
}

type User struct {
	Name string `json:"Name"`
	ID   int64  `json:"ID"`
}

type Event struct {
	Start  time.Time `json:"Start"`
	Finish time.Time `json:"Finish"`
	Owner  User      `json:"Owner"`
	Title  string    `json:"Title"`
	ID     int64     `json:"ID"`
	/*
		Notify: Number of minutes you need to notify before event Start.
		Value always negative.
	*/
	Notify int32 `json:"Notify"`
}

func (e *Event) String() string {
	return fmt.Sprintf(
		"ID: %d\nOwner: %d %s\nTitle: %s\nStart: %s\nFinish: %s\nNotify: %d\n",
		e.ID,
		e.Owner.ID, e.Owner.Name,
		e.Title,
		e.Start.Format(TimeLayout),
		e.Finish.Format(TimeLayout),
		e.Notify,
	)
}

type EventsString struct {
	Events []EventString `json:"Events"`
}

func (e *EventsString) getEvents() Events {
	events := make(Events, len(e.Events))

	for _, es := range e.Events {
		event := es.ToEvent()
		events[event.ID] = event
	}

	return events
}
