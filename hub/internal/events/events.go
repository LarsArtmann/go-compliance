package events

import (
	"context"
	"encoding/json/v2"
	"maps"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/larsartmann/go-datastar"
	"github.com/larsartmann/go-sse"

	"github.com/larsartmann/go-compliance"
)

// SSE event names on the /events stream.
const (
	// EventSnapshot carries the full fleet state: one row per repo. Fresh
	// clients render from it immediately.
	EventSnapshot = "snapshot"
	// EventUpdate carries a state diff: repos added, repos whose displayed
	// state moved, and repos that disappeared.
	EventUpdate = "update"
)

const (
	defaultReplayCapacity   = 256
	defaultSubscriberBuffer = 64
	defaultHeartbeat        = 15 * time.Second
)

// marshalOpts keeps published JSON deterministic (sorted map keys); v2
// otherwise serializes map keys in unspecified order.
//
//nolint:gochecknoglobals // immutable encoding/json/v2 option set, built once
var marshalOpts = json.Deterministic(true)

// Snapshot is the full-state payload of EventSnapshot.
type Snapshot struct {
	GeneratedAt time.Time      `json:"generatedAt"`
	Rows        []RepoRow      `json:"rows"`
	Counts      map[string]int `json:"counts"`
	Total       int            `json:"total"`
}

// Update is the diff payload of EventUpdate.
type Update struct {
	GeneratedAt time.Time      `json:"generatedAt"`
	Added       []RepoRow      `json:"added"`
	Changed     []RepoRow      `json:"changed"`
	Removed     []string       `json:"removed"`
	Counts      map[string]int `json:"counts"`
	Total       int            `json:"total"`
}

// Published is one event as it left the Hub: the wire form (Name, ID, Data)
// plus the typed payload (exactly one of Snapshot or Update is set).
type Published struct {
	Name     string
	ID       string
	Data     string
	Snapshot *Snapshot
	Update   *Update
}

// Hub fans report state out to SSE subscribers with replay support. It owns
// the latest published rows so every surface (JSON stream, DataStar patches)
// reads one source of truth. Safe for concurrent use.
type Hub struct {
	broadcaster      *sse.Broadcaster[Published]
	store            *datastar.MemoryStore
	heartbeat        time.Duration
	subscriberBuffer int
	seq              atomic.Uint64
	dropped          atomic.Int64

	rowsMu sync.RWMutex
	rows   []RepoRow
}

// Option configures a Hub at construction.
type Option func(*Hub)

// WithReplayCapacity bounds how many recent events a reconnecting client
// can still replay.
func WithReplayCapacity(capacity int) Option {
	return func(h *Hub) {
		if capacity > 0 {
			h.store = datastar.NewMemoryStore(capacity)
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

// NewHub returns a Hub with fleet-default replay, buffering, and heartbeat.
func NewHub(opts ...Option) *Hub {
	hub := &Hub{
		store:            datastar.NewMemoryStore(defaultReplayCapacity),
		heartbeat:        defaultHeartbeat,
		subscriberBuffer: defaultSubscriberBuffer,
	}
	for _, opt := range opts {
		opt(hub)
	}

	hub.broadcaster = sse.NewBroadcaster(
		sse.WithOnDrop(func(Published) { hub.dropped.Add(1) }),
		sse.WithBufferSize[Published](hub.subscriberBuffer),
	)

	return hub
}

// Rows returns a copy of the latest published rows; nil before the first
// snapshot. The nil/non-nil split is load-bearing: health checks and stream
// handlers gate "nothing published yet" on nil, while an empty published
// fleet is a valid, ready state.
func (h *Hub) Rows() []RepoRow {
	h.rowsMu.RLock()
	defer h.rowsMu.RUnlock()

	if h.rows == nil {
		return nil
	}

	rows := make([]RepoRow, 0, len(h.rows))
	for _, row := range h.rows {
		row.Severities = copySeverities(row.Severities)
		row.WorstSteps = append([]string(nil), row.WorstSteps...)
		rows = append(rows, row)
	}

	return rows
}

// PublishSnapshot broadcasts the full state and records it for replay.
func (h *Hub) PublishSnapshot(rows []RepoRow) (string, error) {
	snapshot := Snapshot{
		GeneratedAt: time.Now().UTC(),
		Rows:        rows,
		Counts:      CountStates(rows),
		Total:       len(rows),
	}

	h.setRows(rows)

	return h.publish(EventSnapshot, snapshot, &snapshot, nil)
}

// PublishUpdate broadcasts an explicit diff and records it for replay. The
// diff must already describe the transition from the Hub's current rows.
func (h *Hub) PublishUpdate(added, changed []RepoRow, removed []string) (string, error) {
	update := Update{
		GeneratedAt: time.Now().UTC(),
		Added:       added,
		Changed:     changed,
		Removed:     removed,
	}

	rows := h.Rows()
	update.Counts = CountStates(rows)
	update.Total = len(rows)

	return h.publish(EventUpdate, update, nil, &update)
}

// PublishRefresh is the evidence-refresh entry point (the watch/re-derive
// loops call it): it diffs against the Hub's current rows and publishes the
// smallest event that describes the change — a snapshot when no state has
// been published yet (or rows are empty), an update otherwise. Nothing is
// published when nothing changed. Returns whether an event went out.
func (h *Hub) PublishRefresh(next []RepoRow) (bool, error) {
	previous := h.Rows()
	if len(previous) == 0 {
		_, err := h.PublishSnapshot(next)

		return len(next) > 0, err
	}

	added, changed, removed := DiffRows(previous, next)
	if len(added) == 0 && len(changed) == 0 && len(removed) == 0 {
		return false, nil
	}

	h.setRows(next)

	_, err := h.PublishUpdate(added, changed, removed)

	return true, err
}

func (h *Hub) SubscribeFeed() (<-chan Published, func()) {
	feed := h.broadcaster.Subscribe()
	cancel := sync.OnceFunc(func() { h.broadcaster.Unsubscribe(feed) })

	return feed, cancel
}

// Dropped counts events lost to full subscriber buffers since Hub creation.
func (h *Hub) Dropped() int64 { return h.dropped.Load() }

// SubscriberCount reports currently connected subscribers.
func (h *Hub) SubscriberCount() int { return h.broadcaster.SubscriberCount() }

// Shutdown stops broadcasting and waits for subscriber drain or ctx.
func (h *Hub) Shutdown(ctx context.Context) error { return h.broadcaster.Shutdown(ctx) }

// SnapshotJSON renders the hub snapshot payload for rows — the same shape
// the JSON stream's snapshot events carry — as canonical JSON.
func SnapshotJSON(rows []RepoRow) string {
	return mustMarshal(Snapshot{
		GeneratedAt: time.Now().UTC(),
		Rows:        rows,
		Counts:      CountStates(rows),
		Total:       len(rows),
	})
}

// CountStates tallies rows per state name (string form), for stat cards and
// signals.
func CountStates(rows []RepoRow) map[string]int {
	counts := make(map[string]int)
	for _, row := range rows {
		counts[row.State.String()]++
	}

	return counts
}

func (h *Hub) setRows(rows []RepoRow) {
	h.rowsMu.Lock()
	defer h.rowsMu.Unlock()

	h.rows = make([]RepoRow, 0, len(rows))
	for _, row := range rows {
		row.Severities = copySeverities(row.Severities)
		row.WorstSteps = append([]string(nil), row.WorstSteps...)
		h.rows = append(h.rows, row)
	}
}

func (h *Hub) publish(eventName string, payload any, snapshot *Snapshot, update *Update) (string, error) {
	data, err := json.Marshal(payload, marshalOpts)
	if err != nil {
		return "", compliance.NewInternalError(err, "encode %s event", eventName)
	}

	id := h.seq.Add(1)
	event := Published{
		Name:     eventName,
		ID:       strconv.FormatUint(id, 10),
		Data:     string(data),
		Snapshot: snapshot,
		Update:   update,
	}

	h.store.Append(sse.Event{Event: eventName, Data: event.Data, ID: sse.NewEventID(event.ID)})
	h.broadcaster.Broadcast(event)

	return event.ID, nil
}

func copySeverities(severities map[string]int) map[string]int {
	if severities == nil {
		return nil
	}

	copied := make(map[string]int, len(severities))
	maps.Copy(copied, severities)

	return copied
}

func mustMarshal(payload any) string {
	data, err := json.Marshal(payload, marshalOpts)
	if err != nil {
		return "{}"
	}

	return string(data)
}
