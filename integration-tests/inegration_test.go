package integration_tests

import (
	j "encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

const day = 24 * time.Hour

func TestCalendar(t *testing.T) {
	var id int64 = 0
	user := User{
		Name: "newUser",
		ID:   1,
	}

	t.Run("create_event", func(t *testing.T) {
		s, err := time.Parse(TimeLayout, "2022-01-02T15:04:00Z")
		require.NoError(t, err)

		f, err := time.Parse(TimeLayout, "2022-02-02T15:04:00Z")
		require.NoError(t, err)

		id++
		event1 := Event{
			ID:     id,
			Owner:  user,
			Title:  "myNewTitle",
			Start:  processTime(s),
			Finish: processTime(f),
			Notify: 120,
		}
		createEvent(t, event1)

		id++
		event2 := Event{
			ID:     id,
			Owner:  user,
			Title:  "some text",
			Start:  processTime(time.Now().Add(-5 * time.Hour)), // Start must be before the
			Finish: processTime(time.Now()),
			Notify: 150,
		}
		createEvent(t, event2)

		e := listUserEvents(t, user.Name, 2)
		require.Equal(t, 2, len(e))
		require.Equal(t, event1, e[event1.ID])
		require.Equal(t, event2, e[event2.ID])
	})

	t.Run("update_event", func(t *testing.T) {
		updatedEvent := Event{
			ID:     id, // updating last created event
			Owner:  user,
			Title:  "some another text",
			Start:  processTime(time.Now().Add(-10 * time.Hour)),
			Finish: processTime(time.Now().Add(15 * time.Hour)),
			Notify: 180,
		}

		resp, err := http.Get(addr + fmt.Sprintf(
			"/update/%d/%s/%q/%q/%d",
			updatedEvent.ID,
			updatedEvent.Title,
			updatedEvent.Start.Format(TimeLayout), // Start must be before the
			updatedEvent.Finish.Format(TimeLayout),
			updatedEvent.Notify,
		))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		events := listUserEvents(t, user.Name, 2)
		require.Equal(t, updatedEvent, events[updatedEvent.ID])
	})

	t.Run("delete_event", func(t *testing.T) {
		updatedEvent := Event{
			ID: 1,
		}

		resp, err := http.Get(addr + fmt.Sprintf(
			"/delete/%d",
			updatedEvent.ID,
		))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		events := listUserEvents(t, user.Name, 1)
		_, found := events[updatedEvent.ID]
		require.False(t, false, found)

		resp, err = http.Get(addr + fmt.Sprintf(
			"/delete/%d",
			id,
		))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		listUserEvents(t, user.Name, 0)
	})

	user = User{ID: 2, Name: "SecondUser"}
	t.Run("list_upcoming_day", func(t *testing.T) {
		id++
		eDay := Event{
			ID:     id,
			Owner:  user,
			Title:  "string",
			Start:  processTime(time.Now().Add(25 * time.Hour)),
			Finish: processTime(time.Now().Add(2 * 365 * day)),
			Notify: 120,
			// Event will be started in 25 hours, but
			// we need to notify the user
			// 2 hours before the start of the event
		}
		createEvent(t, eDay)

		id++
		eWeek := eDay
		eWeek.ID = id
		eWeek.Start = processTime(time.Now().Add(7 * day))
		eWeek.Notify = 24 * 60
		createEvent(t, eWeek)

		id++
		eMonth := eWeek
		eMonth.ID = id
		eMonth.Start = processTime(time.Now().Add(25 * day))
		createEvent(t, eMonth)

		id++
		eYear := eMonth
		eYear.ID = id
		eYear.Start = processTime(time.Now().Add(365 * day))
		createEvent(t, eYear)

		resp, err := http.Get(addr + fmt.Sprintf(
			"/list/%s/%d",
			eDay.Owner.Name,
			0,
		))
		require.NoError(t, err)
		events := getEvents(t, resp)
		require.Equal(t, 1, len(events))
		_, found := events[eDay.ID]
		require.True(t, found)

		resp, err = http.Get(addr + fmt.Sprintf(
			"/list/%s/%d",
			eDay.Owner.Name,
			1,
		))
		require.NoError(t, err)
		events = getEvents(t, resp)
		require.Equal(t, 2, len(events))
		_, found = events[eDay.ID]
		require.True(t, found)
		_, found = events[eWeek.ID]
		require.True(t, found)

		resp, err = http.Get(addr + fmt.Sprintf(
			"/list/%s/%d",
			eDay.Owner.Name,
			2,
		))
		require.NoError(t, err)
		events = getEvents(t, resp)
		require.Equal(t, 3, len(events))
		_, found = events[eDay.ID]
		require.True(t, found)
		_, found = events[eWeek.ID]
		require.True(t, found)
		_, found = events[eMonth.ID]
		require.True(t, found)

		resp, err = http.Get(addr + fmt.Sprintf(
			"/list/%s/%d",
			eDay.Owner.Name,
			8, // 0 - day, 1 - week, 2 - month, day by default
		))
		require.NoError(t, err)
		events = getEvents(t, resp)
		require.Equal(t, 1, len(events))
		_, found = events[eDay.ID]
		require.True(t, found)
	})
}

func processTime(t time.Time) time.Time {
	// SERVER ROUNDS TIME TO MINUTES AND SETS LOCAL TO UTC
	return t.Round(time.Minute).UTC()
}

func listUserEvents(t *testing.T, username string, l int) Events {
	resp, err := http.Get(addr + "/list/" + username)
	require.NoError(t, err)
	events := getEvents(t, resp)
	require.Len(t, events, l)
	return events
}

func getEvents(t *testing.T, resp *http.Response) Events {
	json, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Nil(t, resp.Body.Close())

	eventQ := EventsString{}
	err = j.Unmarshal(json, &eventQ)
	require.NoError(t, err)

	return eventQ.getEvents()
}

func createEvent(t *testing.T, e Event) {
	resp, err := http.Get(addr + fmt.Sprintf(
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
