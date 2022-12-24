package integration_tests

import (
	j "encoding/json"
	"fmt"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

var addrGRPC = "http://" + net.JoinHostPort(os.Getenv("API_HOST"), os.Getenv("GRPC_PORT"))

func TestGRPCEndpoint(t *testing.T) {
	user := createUser("newUserGRPC")

	t.Run("create_event", func(t *testing.T) {
		s, err := time.Parse(TimeLayout, "2022-01-02T15:04:00Z")
		require.NoError(t, err)

		f, err := time.Parse(TimeLayout, "2022-02-02T15:04:00Z")
		require.NoError(t, err)

		event1 := storage.Event{
			Owner:  user,
			Title:  "myNewTitle",
			Start:  processTime(s),
			Finish: processTime(f),
			Notify: 120,
		}
		createEventGRPC(t, &event1)

		event2 := storage.Event{
			Owner:  user,
			Title:  "some text",
			Start:  processTime(time.Now().Add(-5 * time.Hour)), // Start must be before the
			Finish: processTime(time.Now()),
			Notify: 150,
		}
		createEventGRPC(t, &event2)

		events, err := checkUserEventsGRPC(user.Name)
		require.NoError(t, err)
		require.Len(t, events, 2)
		require.Equal(t, event1, events[event1.ID])
		require.Equal(t, event2, events[event2.ID])
	})

	t.Run("update_event", func(t *testing.T) {
		eventToUpdate := storage.Event{
			ID:     eventID, // updating last created event
			Owner:  user,
			Title:  "some another text",
			Start:  processTime(time.Now().Add(-10 * time.Hour)),
			Finish: processTime(time.Now().Add(15 * time.Hour)),
			Notify: 180,
		}

		resp, err := http.Get(addrGRPC + fmt.Sprintf(
			"/update/%d/%s/%q/%q/%d",
			eventToUpdate.ID,
			eventToUpdate.Title,
			eventToUpdate.Start.Format(TimeLayout), // Start must be before the
			eventToUpdate.Finish.Format(TimeLayout),
			eventToUpdate.Notify,
		))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		events, err := checkUserEventsGRPC(user.Name)
		require.NoError(t, err)
		require.Len(t, events, 2)
		require.Equal(t, eventToUpdate, events[eventToUpdate.ID])
	})

	t.Run("delete_event", func(t *testing.T) {
		updatedEvent := storage.Event{
			ID: eventID,
		}

		resp, err := http.Get(addrGRPC + fmt.Sprintf(
			"/delete/%d",
			updatedEvent.ID,
		))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		events, err := checkUserEventsGRPC(user.Name)
		require.NoError(t, err)
		require.Len(t, events, 1)

		_, found := events[updatedEvent.ID]
		require.False(t, found)

		resp, err = http.Get(addrGRPC + fmt.Sprintf(
			"/delete/%d",
			eventID,
		))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		events, err = checkUserEventsGRPC(user.Name)
		require.NoError(t, err)
		require.Len(t, events, 1)
	})

	user = createUser("SecondGRPCUser")
	t.Run("list_upcoming_day", func(t *testing.T) {
		eDay := storage.Event{
			ID:     eventID,
			Owner:  user,
			Title:  "string",
			Start:  processTime(time.Now().Add(25 * time.Hour)),
			Finish: processTime(time.Now().Add(2 * 365 * day)),
			Notify: 120,
			// Event will be started in 25 hours, but
			// we need to notify the user
			// 2 hours before the start of the event
		}
		createEventGRPC(t, &eDay)

		eWeek := eDay
		eWeek.ID = eventID
		eWeek.Start = processTime(time.Now().Add(7 * day))
		eWeek.Notify = 24 * 60
		createEventGRPC(t, &eWeek)

		eMonth := eWeek
		eMonth.ID = eventID
		eMonth.Start = processTime(time.Now().Add(25 * day))
		createEventGRPC(t, &eMonth)

		eYear := eMonth
		eYear.ID = eventID
		eYear.Start = processTime(time.Now().Add(365 * day))
		createEventGRPC(t, &eYear)

		resp, err := http.Get(addrGRPC + fmt.Sprintf(
			"/list/%s/%d",
			eDay.Owner.Name,
			0,
		))
		require.NoError(t, err)
		events, err := getEvents(resp)
		require.NoError(t, err)
		require.Equal(t, 1, len(events))
		_, found := events[eDay.ID]
		require.True(t, found)

		resp, err = http.Get(addrGRPC + fmt.Sprintf(
			"/list/%s/%d",
			eDay.Owner.Name,
			1,
		))
		require.NoError(t, err)
		events, err = getEvents(resp)
		require.NoError(t, err)
		require.Equal(t, 2, len(events))
		_, found = events[eDay.ID]
		require.True(t, found)
		_, found = events[eWeek.ID]
		require.True(t, found)

		resp, err = http.Get(addrGRPC + fmt.Sprintf(
			"/list/%s/%d",
			eDay.Owner.Name,
			2,
		))
		require.NoError(t, err)
		events, err = getEvents(resp)
		require.NoError(t, err)
		require.Len(t, events, 3)

		_, found = events[eDay.ID]
		require.True(t, found)
		_, found = events[eWeek.ID]
		require.True(t, found)
		_, found = events[eMonth.ID]
		require.True(t, found)

		expected := grpcError{
			Code:    3,
			Message: "type mismatch, parameter: Until, error: 8 is not valid",
			Details: []string{},
		}

		current := grpcError{}

		resp, err = http.Get(addrGRPC + fmt.Sprintf(
			"/list/%s/%d",
			eDay.Owner.Name,
			8, // 0 - day, 1 - week, 2 - month, other should return error
		))
		require.NoError(t, err)
		json, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Nil(t, resp.Body.Close())

		err = j.Unmarshal(json, &current)
		require.NoError(t, err)
		require.Equal(t, expected, current)
	})
}

func checkUserEventsGRPC(username string) (EventMap, error) {
	resp, err := http.Get(addrGRPC + "/list/" + username)
	if err != nil {
		return nil, err
	}

	m, err := getEvents(resp)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func createEventGRPC(t *testing.T, e *storage.Event) {
	eventID++
	e.ID = eventID
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
