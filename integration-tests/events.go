package integration_tests

import (
	j "encoding/json"
	"fmt"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

const TimeLayout = "2006-01-02T15:04:05Z07:00"

var addrGRPC = "http://" + net.JoinHostPort(os.Getenv("API_HOST"), os.Getenv("GRPC_PORT"))

//var addrHTTP = "http://" + net.JoinHostPort(os.Getenv("API_HOST"), os.Getenv("HTTP_PORT"))

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

func getEvents(t *testing.T, resp *http.Response) EventMap {
	json, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Nil(t, resp.Body.Close())

	eventQ := EventsString{}
	err = j.Unmarshal(json, &eventQ)

	log.Println("UNMARSHALED")
	log.Println(string(json))
	log.Println()
	log.Println(eventQ)
	log.Println()

	require.NoError(t, err)

	return eventQ.getEvents()
}

func createEvent(t *testing.T, e storage.Event) {
	resp, err := http.Get(addrGRPC + fmt.Sprintf(
		"/create/%s/%s/%q/%q/%d",
		e.Owner.Name,
		e.Title,
		e.Start.Format(TimeLayout),
		e.Finish.Format(TimeLayout),
		e.Notify,
	))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}
