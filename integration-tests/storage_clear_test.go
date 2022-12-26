package integration_tests

import (
	"github.com/ennwy/calendar/internal/notification"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStorageClear(t *testing.T) {
	user := createUser("Alex")

	event := storage.Event{
		Owner: user,
		Title: "some interesting description",
		// started two years ago
		Start: processTime(time.Now()).Add(-2 * 365 * 24 * time.Hour),
		// finished one year and one day ago
		Finish: processTime(time.Now()).Add(-1 * 366 * 24 * time.Hour),
		Notify: 11,
	}

	err := createEventGRPC(&event)
	require.NoError(t, err)

	events, err := checkUserEventsGRPC(event.Owner.Name)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, event, events[event.ID])

	p := notification.Period{}
	require.NoError(t, p.Set())
	time.Sleep(p.Clear)

	events, err = checkUserEventsGRPC(event.Owner.Name)
	require.NoError(t, err)
	require.Len(t, events, 0)
}
