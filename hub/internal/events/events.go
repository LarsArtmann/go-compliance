package events

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/larsartmann/go-sse"
)

// SSE event names on the /events stream.
const (
	// EventSnapshot carries the full fleet state: one row per repo. Fresh
	// clients render from it immediately.
	EventSnapshot = "snapshot"
	// EventUpdate carries a state diff: only repos whose displayed state
	// moved plus repos that disappeared.
	EventUpdate = "update"
)

const (
	defaultReplayCapacity   = 256
	defaultSubscriberBuffer = 64
	defaultHeartbeat        = 15 * time.Second
)

// Snapshot is the full-state payload of EventSnapshot.
type Snapshot struct {
	GeneratedAt time.Time `json:"generatedAt"`
	Rows        []RepoRow `json:"rows"`
}

// Update is the diff payload of EventUpdate.
type Update struct {
	GeneratedAt time.Time `json:"generatedAt"`
	Changed     []RepoRow `json:"changed"`
	Removed     []string  `json:"removed"`
}

// Hub fans report state out to SSE subscribers with replay support. It is
// safe for concurrent use; Close is owned by Shutdown.
type Hub struct {
	broadcaster      *sse.Broadcaster[sse.Event]
	store            *ringStore
	heartbeat        time.Duration
	subscriberBuffer int
	seq              atomic.Uint64
	dropped          atomic.Int64
}

// Option configures a Hub at construction.
type Option func(*Hub)

// WithReplayCapacity bounds how many recent events a reconnecting client
// can still replay.
func WithReplayCapacity(capacity int) Option {
	return func(h *Hub) {
		if capacity > 0 {
			h.store = &ringStore{cap: capacity}
		}
	}
}

// WithSubscriberBuffer sets each subscriber's channel buffer before drops
// start counting.
func WithSubscriberBuffer(size int) Option {
	return func(h *Hub) {
		if size > 0 {
			h.subscriberBuffer = size
		}
	}
}

// WithHeartbeat sets the SSE comment heartbeat interval.
func WithHeartbeat(interval time.Duration) Option {
	return func(h *Hub) {
		if interval > 0 {
			h.heartbeat = interval
		}
	}
}

var marshalOpts = json.Deterministic(true)

// NewHub returns a Hub with fleet-default replay, buffering, and heartbeat.
func NewHub(opts ...Option) *Hub {
	hub := &Hub{
		store:            &ringStore{cap: defaultReplayCapacity},
		heartbeat:        defaultHeartbeat,
		subscriberBuffer: defaultSubscriberBuffer,
	}
	for _, opt := range opts {
		opt(hub)
	}

	hub.broadcaster = sse.NewBroadcaster[sse.Event](
		sse.WithOnDrop(func(sse.Event) { hub.dropped.Add(1) }),
		sse.WithBufferSize[sse.Event](hub.subscriberBuffer),
	)

	return hub
}

// PublishSnapshot broadcasts the full state to all subscribers and records
// it for replay. Returns the minted event ID.
func (h *Hub) PublishSnapshot(rows []RepoRow) (string, error) {
	return h.publish(EventSnapshot, Snapshot{GeneratedAt: time.Now().UTC(), Rows: rows})
}

// PublishUpdate broadcasts a state diff and records it for replay.
func (h *Hub) PublishUpdate(changed []RepoRow, removed []string) (string, error) {
	return h.publish(EventUpdate, Update{GeneratedAt: time.Now().UTC(), Changed: changed, Removed: removed})
}

// PublishRefresh is the evidence-refresh entry point (the watch/re-derive
// loops call it): the first refresh publishes a snapshot, later ones publish
// only the diff. Nothing is published when nothing changed. Returns whether
// an event went out.
func (h *Hub) PublishRefresh(previous, next []RepoRow) (bool, error) {
	if len(previous) == 0 {
		_, err := h.PublishSnapshot(next)
		return len(next) > 0, err
	}

	changed, removed := DiffRows(previous, next)
	if len(changed) == 0 && len(removed) == 0 {
		return false, nil
	}

	_, err := h.PublishUpdate(changed, removed)
	return true, err
}

// Handler serves the SSE stream: replayed events after a Last-Event-ID, the
// current snapshot (so fresh clients render immediately), then live events
// with a comment heartbeat. The current callback runs on every connection;
// it must be safe for concurrent reads.
func (h *Hub) Handler(current func() []RepoRow) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		if lastID := sse.LastEventIDFromRequest(r); !lastID.IsZero() {
			if _, err := sse.Replay(stream, h.store, lastID); err != nil {
				return
			}
		}

		if rows := current(); rows != nil {
			_ = stream.Send(sse.Event{
				Event: EventSnapshot,
				Data:  mustMarshal(Snapshot{GeneratedAt: time.Now().UTC(), Rows: rows}),
			})
		}

		subscription := h.broadcaster.Subscribe()
		defer h.broadcaster.Unsubscribe(subscription)

		go stream.Heartbeat(stream.Context(), h.heartbeat)

		for {
			select {
			case <-stream.Context().Done():
				return
			case evt, ok := <-subscription:
				if !ok || stream.Send(evt) != nil {
					return
				}
			}
		}
	})
}

// Dropped counts events lost to full subscriber buffers since Hub creation.
func (h *Hub) Dropped() int64 { return h.dropped.Load() }

// SubscriberCount reports currently connected subscribers.
func (h *Hub) SubscriberCount() int { return h.broadcaster.SubscriberCount() }

// Shutdown stops broadcasting and waits for subscriber drain or ctx.
func (h *Hub) Shutdown(ctx context.Context) error { return h.broadcaster.Shutdown(ctx) }

func (h *Hub) publish(eventName string, payload any) (string, error) {
	data, err := json.Marshal(payload, marshalOpts)
	if err != nil {
		return "", err
	}

	id := h.seq.Add(1)
	event := sse.Event{
		Event: eventName,
		Data:  string(data),
		ID:    sse.NewEventID(strconv.FormatUint(id, 10)),
	}

	h.store.append(event)
	h.broadcaster.Broadcast(event)

	return event.ID.Get(), nil
}

func mustMarshal(payload any) string {
	data, err := json.Marshal(payload, marshalOpts)
	if err != nil {
		return "{}"
	}

	return string(data)
}
