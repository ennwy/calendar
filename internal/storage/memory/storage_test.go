package memstorage

import (
	"context"
	"sync"
	"testing"

	"github.com/ennwy/calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestBasicLogic(t *testing.T) {
	t.Run("basic_logic_test", func(t *testing.T) {
		s := New()
		testEvent1 := storage.Event{Title: "some text", OwnerID: 1}
		testEvent2 := storage.Event{Title: "another text", OwnerID: 2}
		_ = s.CreateEvent(ctx, &testEvent1)
		_ = s.CreateEvent(ctx, &testEvent2)

		eventList1, _ := s.ListEvents(ctx, 1)
		require.Equal(t, 1, len(eventList1))

		eventList2, _ := s.ListEvents(ctx, 2)
		require.Equal(t, 1, len(eventList2))
		require.NotEqual(t, eventList1, eventList2)

		allEvents := GlobalList(s, testEvent1.OwnerID, testEvent2.OwnerID)
		require.Equal(t, 2, len(allEvents))

		_ = s.DeleteEvent(ctx, testEvent1.ID)
		eventList1, _ = s.ListEvents(ctx, testEvent1.OwnerID)

		require.Equal(t, 0, len(eventList1))
		require.Equal(t, 1, len(eventList2))

		testEvent2.Title = "another title"
		_ = s.UpdateEvent(ctx, testEvent2)
		eventList2, _ = s.ListEvents(ctx, testEvent2.OwnerID)

		require.Equal(t, 1, len(eventList2))
		require.Equal(t, testEvent2.OwnerID, eventList2[0].OwnerID)
		require.Equal(t, testEvent2.Title, eventList2[0].Title)

		_ = s.DeleteEvent(ctx, testEvent2.ID)
		eventList2, _ = s.ListEvents(ctx, testEvent2.OwnerID)
		require.Equal(t, 0, len(eventList2))
	})
}

func TestAsync(t *testing.T) {
	s := New()
	var eventList []storage.Event
	wg := &sync.WaitGroup{}
	baseTittle := "title"

	ownerIDs := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	t.Run("create_event_async", func(t *testing.T) {
		for _, ownerID := range ownerIDs {
			wg.Add(1)
			event := storage.Event{
				OwnerID: ownerID,
				Title:   baseTittle,
			}

			go func(event storage.Event, wg *sync.WaitGroup) {
				_ = s.CreateEvent(ctx, &event)
				wg.Done()
			}(event, wg)
		}

		wg.Wait()
		eventList = GlobalList(s, ownerIDs...)
		require.Equal(t, len(ownerIDs), len(eventList))
	})

	t.Run("delete_event_async", func(t *testing.T) {
		// event indexing starts from 1, so add 1 to result
		// deleting last 5 events
		for _, id := range ownerIDs[5:] {

			wg.Add(1)
			go func(id int64, wg *sync.WaitGroup) {
				e, _ := s.ListEvents(ctx, id)
				_ = s.DeleteEvent(ctx, e[0].ID)

				wg.Done()
			}(id, wg)
		}

		wg.Wait()
		eventList = GlobalList(s, ownerIDs...)
		require.Equal(t, 5, len(eventList))

		for _, event := range eventList {
			require.NotContains(t, []int64{6, 7, 8, 9, 10}, event.OwnerID)
		}
	})

	t.Run("update_event_async", func(t *testing.T) {
		// updating first 5 events
		eventsToUpdate := []int64{0, 1, 2, 3, 4}

		allEvents := GlobalList(s, eventsToUpdate...)
		require.Equal(t, 5, len(allEvents))

		for i, e := range allEvents {
			wg.Add(1)
			event := storage.Event{
				ID:      e.ID,
				OwnerID: e.OwnerID,
				// a, b, c...
				Title: string(rune(i + 97)),
			}

			go func(event storage.Event, wg *sync.WaitGroup) {
				_ = s.UpdateEvent(ctx, event)
				wg.Done()
			}(event, wg)
		}

		wg.Wait()
		eventList = GlobalList(s, eventsToUpdate...)
		require.Equal(t, 5, len(eventList))

		for _, event := range eventList {
			require.Contains(t, []string{"a", "b", "c", "d", "e"}, event.Title)
		}
	})

	t.Run("list_events_async", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func(i int64) {
				_, _ = s.ListEvents(ctx, i)
				wg.Done()
			}(int64(i))
		}

		wg.Wait()
	})
}

func GlobalList(s *Storage, ownerIDs ...int64) []storage.Event {
	allEvents := make([]storage.Event, 0)
	for _, id := range ownerIDs {
		list, _ := s.ListEvents(context.Background(), id)
		allEvents = append(allEvents, list...)
	}

	return allEvents
}
