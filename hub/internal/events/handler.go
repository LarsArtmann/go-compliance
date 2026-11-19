package events

import (
	"net/http"
	"time"

	"github.com/larsartmann/go-sse"
)

// Handler serves the JSON SSE stream: replayed events after a
// Last-Event-ID, the current snapshot (so fresh clients render
// immediately), then live events with a comment heartbeat.

// Handler serves the JSON SSE stream: replayed events after a
// Last-Event-ID, the current snapshot (so fresh clients render
// immediately), then live events with a comment heartbeat.
func (h *Hub) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		if lastID := sse.LastEventIDFromRequest(r); !lastID.IsZero() {
			if _, err := sse.Replay(stream, h.store, lastID); err != nil {
				return
			}
		}

		if rows := h.Rows(); rows != nil {
			snapshot := Snapshot{
				GeneratedAt: time.Now().UTC(),
				Rows:        rows,
				Counts:      CountStates(rows),
				Total:       len(rows),
			}

			_ = stream.Send(sse.Event{
				Event: EventSnapshot,
				Data:  mustMarshal(snapshot),
			})
		}

		feed, cancel := h.SubscribeFeed()
		defer cancel()

		//nolint:contextcheck // stream.Context() IS the request context, hidden behind the sse stream API
		go stream.Heartbeat(stream.Context(), h.heartbeat)

		for {
			select {
			case <-stream.Context().Done():
				return
			case published, ok := <-feed:
				if !ok {
					return
				}

				if stream.Send(
					sse.Event{Event: published.Name, Data: published.Data, ID: sse.NewEventID(published.ID)},
				) != nil {
					return
				}
			}
		}
	})
}

// SubscribeFeed returns a live typed feed of published events plus a cancel
// function. Callers must drain the feed or cancel promptly: Shutdown waits
// for subscriber drain.
