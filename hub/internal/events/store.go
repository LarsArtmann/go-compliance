package events

import (
	"strconv"
	"sync"

	"github.com/larsartmann/go-sse"
)

// ringStore is an in-memory sse.EventStore keeping the newest capacity
// events for Last-Event-ID replay. IDs are monotonic integers minted by the
// Hub, so "strictly after" is a numeric comparison and eviction of old
// events only shrinks the replayable window.
type ringStore struct {
	mu     sync.Mutex
	events []sse.Event
	cap    int
}

// EventsAfter implements sse.EventStore: events with numeric IDs strictly
// after lastID, in ascending order. A zero or unparseable lastID replays
// everything the ring still holds.
func (s *ringStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := uint64(0)
	if parsed, ok := parseSeq(lastID); ok {
		cutoff = parsed
	}

	var replayable []sse.Event
	for _, evt := range s.events {
		if seq, ok := parseSeq(evt.ID); ok && seq > cutoff {
			replayable = append(replayable, evt)
		}
	}

	return replayable, nil
}

func (s *ringStore) append(evt sse.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, evt)
	if len(s.events) > s.cap {
		s.events = s.events[len(s.events)-s.cap:]
	}
}

func parseSeq(id sse.EventID) (uint64, bool) {
	if id.IsZero() {
		return 0, false
	}

	parsed, err := strconv.ParseUint(id.Get(), 10, 64)
	if err != nil {
		return 0, false
	}

	return parsed, true
}
