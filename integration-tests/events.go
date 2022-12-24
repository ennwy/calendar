package integration_tests

import (
	j "encoding/json"
	"github.com/ennwy/calendar/internal/storage"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	//TimeLayout = "2006-01-02T15:04:05Z07:00"
	TimeLayout = time.RFC3339Nano

	day = 24 * time.Hour
)

var userID, eventID int64

// JSON convert int64 to string, so we need to parse it to int64

type EventMap map[int64]storage.Event

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

func (e *EventString) ToEvent() storage.Event {
	// WE CAN'T GET ERROR HERE BECAUSE SERVER RETURNS ONLY INT VALUES
	ownerID, _ := strconv.ParseInt(e.Owner.ID, 10, 64)
	eventID, _ := strconv.ParseInt(e.ID, 10, 64)

	return storage.Event{
		ID: eventID,
		Owner: storage.User{
			ID:   ownerID,
			Name: e.Owner.Name,
		},
		Title:  e.Title,
		Start:  e.Start,
		Finish: e.Finish,
		Notify: e.Notify,
	}
}

type EventsString struct {
	Events []EventString `json:"Events"`
}

func (e *EventsString) getEvents() EventMap {
	events := make(EventMap, len(e.Events))

	for _, es := range e.Events {
		event := es.ToEvent()
		events[event.ID] = event
	}

	return events
}

type grpcError struct {
	Code    int32    `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details"`
}

func processTime(t time.Time) time.Time {
	// SERVER ROUNDS TIME TO MINUTES AND SETS LOCAL TO UTC
	return t.Round(time.Minute).UTC()
}

func createUser(username string) storage.User {
	userID++
	user := storage.User{
		Name: username,
		ID:   userID,
	}

	return user
}

func getEvents(resp *http.Response) (EventMap, error) {
	json, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = resp.Body.Close(); err != nil {
		return nil, err
	}

	eventQ := EventsString{}
	err = j.Unmarshal(json, &eventQ)

	log.Println("UNMARSHALED")
	log.Println(string(json))
	log.Println()
	log.Println(eventQ)
	log.Println()

	if err != nil {
		return nil, err
	}

	return eventQ.getEvents(), nil
}
