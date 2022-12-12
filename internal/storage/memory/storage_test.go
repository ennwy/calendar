package memstorage

import (
	"context"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/stretchr/testify/require"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
)

var ctx = context.Background()

func TestBasicLogic(t *testing.T) {
	t.Run("basic_logic_test", func(t *testing.T) {
		s := New()

		username1 := "1"
		username2 := "2"
		testEvent1 := storage.Event{Title: "some text", Owner: storage.User{ID: 1, Name: username1}}
		testEvent2 := storage.Event{Title: "another text", Owner: storage.User{ID: 2, Name: username2}}
		_ = s.CreateEvent(ctx, &testEvent1)
		_ = s.CreateEvent(ctx, &testEvent2)

		eventList1, _ := s.ListUserEvents(ctx, username1)
		require.Equal(t, 1, len(eventList1))

		eventList2, _ := s.ListUserEvents(ctx, username2)
		require.Equal(t, 1, len(eventList2))
		require.NotEqual(t, eventList1, eventList2)

		allEvents := getUsersEvents(s, testEvent1.Owner.Name, testEvent2.Owner.Name)
		require.Equal(t, 2, len(allEvents))

		_ = s.DeleteEvent(ctx, testEvent1.ID)
		eventList1, _ = s.ListUserEvents(ctx, testEvent1.Owner.Name)

		require.Equal(t, 0, len(eventList1))
		require.Equal(t, 1, len(eventList2))

		testEvent2.Title = "another title"
		_ = s.UpdateEvent(ctx, &testEvent2)
		eventList2, _ = s.ListUserEvents(ctx, testEvent2.Owner.Name)

		require.Equal(t, 1, len(eventList2))
		require.Equal(t, testEvent2.Owner.ID, eventList2[0].Owner.ID)
		require.Equal(t, testEvent2.Title, eventList2[0].Title)

		_ = s.DeleteEvent(ctx, testEvent2.ID)
		eventList2, _ = s.ListUserEvents(ctx, testEvent2.Owner.Name)
		require.Equal(t, 0, len(eventList2))
	})
}

func TestAsync(t *testing.T) {
	s := New()
	var eventList []storage.Event
	wg := &sync.WaitGroup{}
	baseTittle := "title"

	ownerNames := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	t.Run("create_event_async", func(t *testing.T) {
		for _, name := range ownerNames {

			event := storage.Event{
				Owner: storage.User{
					Name: name,
				},
				Title: baseTittle,
			}
			wg.Add(1)
			go func(event storage.Event, wg *sync.WaitGroup) {
				_ = s.CreateEvent(ctx, &event)
				wg.Done()
			}(event, wg)
		}

		wg.Wait()
		eventList = getUsersEvents(s, ownerNames...)
		require.Equal(t, len(ownerNames), len(eventList))
	})

	t.Run("delete_event_async", func(t *testing.T) {
		// event indexing starts from 1, so add 1 to result
		// deleting last 5 events
		for _, name := range ownerNames[5:] {

			wg.Add(1)
			go func(name string, wg *sync.WaitGroup) {
				e, _ := s.ListUserEvents(ctx, name)
				_ = s.DeleteEvent(ctx, e[0].ID)

				wg.Done()
			}(name, wg)
		}

		wg.Wait()
		eventList = getUsersEvents(s, ownerNames...)
		require.Equal(t, 5, len(eventList))

		for _, event := range eventList {
			require.NotContains(t, []string{"5", "6", "7", "8", "9"}, event.Owner.Name)
			log.Println(event.Owner.Name)
		}
	})

	t.Run("update_event_async", func(t *testing.T) {
		// updating first 5 events
		eventsToUpdate := []string{"0", "1", "2", "3", "4"}

		allEvents := getUsersEvents(s, eventsToUpdate...)
		require.Equal(t, 5, len(allEvents))

		for i, e := range allEvents {
			wg.Add(1)
			event := storage.Event{
				ID: e.ID,
				Owner: storage.User{
					ID:   e.Owner.ID,
					Name: e.Owner.Name,
				},
				// a, b, c...
				Title: eventsToUpdate[i],
			}

			go func(event storage.Event, wg *sync.WaitGroup) {
				_ = s.UpdateEvent(ctx, &event)
				wg.Done()
			}(event, wg)
		}

		wg.Wait()
		eventList = getUsersEvents(s, eventsToUpdate...)
		require.Equal(t, 5, len(eventList))

		for _, event := range eventList {
			require.Contains(t, []string{"0", "1", "2", "3", "4"}, event.Title)
		}
	})

	t.Run("list_events_async", func(t *testing.T) {
		for i := int64(0); i < 100; i++ {
			name := strconv.FormatInt(i, 10)
			wg.Add(1)
			go func(name string) {
				_, _ = s.ListUserEvents(ctx, name)
				wg.Done()
			}(name)
		}

		wg.Wait()
	})

	t.Run("async_update_same_event", func(t *testing.T) {
		times := 50
		events := make([]*storage.Event, 0, times)

		for i := 0; i < times; i++ {
			name := strconv.FormatInt(int64(i), 10)

			event := storage.Event{
				Owner: storage.User{
					ID:   int64(i),
					Name: "event" + name,
				},
			}

			err := s.CreateEvent(ctx, &event)
			require.NoError(t, err)

			events = append(events, &event)
		}

		wg := &sync.WaitGroup{}
		for i, event := range events {
			event.Title = strconv.FormatInt(int64(i), 10)

			wg.Add(1)
			go func(e *storage.Event, wg *sync.WaitGroup) {
				err := s.UpdateEvent(ctx, e)
				require.NoError(t, err)

				wg.Done()
			}(event, wg)
		}

		wg.Wait()
	})

	t.Run("update_data_race", func(t *testing.T) {
		event := &storage.Event{
			Owner: storage.User{
				ID:   100600,
				Name: "John",
			},
			Title: "some string",
		}

		err := s.CreateEvent(ctx, event)
		require.NoError(t, err)

		wg := &sync.WaitGroup{}
		for i := 0; i < 50; i++ {
			title := strconv.FormatInt(int64(i), 10)

			s.mu.Lock()
			event.Title = title
			s.mu.Unlock()

			wg.Add(1)

			go func(e *storage.Event, wg *sync.WaitGroup) {
				err := s.UpdateEvent(ctx, e)
				require.NoError(t, err)
				wg.Done()
			}(event, wg)
		}

		wg.Wait()

	})
	t.Run("clear", func(t *testing.T) {
		username := "New User"
		event1 := &storage.Event{
			Owner: storage.User{
				ID:   500,
				Name: username,
			},
			Finish: time.Now().Add(time.Hour * 24 * -370),
		}
		event2 := &storage.Event{
			Owner: storage.User{
				ID:   500,
				Name: username,
			},
			Finish: time.Now().Add(time.Hour * 24 * -355),
		}

		err := s.CreateEvent(ctx, event1)
		require.NoError(t, err)

		err = s.CreateEvent(ctx, event2)
		require.NoError(t, err)

		events, err := s.ListUserEvents(ctx, username)
		require.NoError(t, err)
		require.Equal(t, 2, len(events))

		err = s.Clean(ctx, time.Hour*24*365)
		require.NoError(t, err)

		events, err = s.ListUserEvents(ctx, username)
		require.NoError(t, err)
		require.Equal(t, 1, len(events))
	})
}

func getUsersEvents(s *Storage, usernames ...string) []storage.Event {
	allEvents := make([]storage.Event, 0, 1)
	for _, id := range usernames {
		list, _ := s.ListUserEvents(context.Background(), id)
		allEvents = append(allEvents, list...)
	}

	return allEvents
}
