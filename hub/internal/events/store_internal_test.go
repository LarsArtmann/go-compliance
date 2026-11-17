package events

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-sse"
)

var _ = Describe("replay ring store", func() {
	newEvent := func(id string) sse.Event { return sse.Event{ID: sse.NewEventID(id)} }

	It("returns events strictly after the given ID, ascending", func() {
		store := &ringStore{cap: 8}
		for _, id := range []string{"1", "2", "3"} {
			store.append(newEvent(id))
		}

		replayable, err := store.EventsAfter(sse.NewEventID("1"))
		Expect(err).NotTo(HaveOccurred())
		Expect(replayable).To(HaveLen(2))
		Expect(replayable[0].ID.Get()).To(Equal("2"))
		Expect(replayable[1].ID.Get()).To(Equal("3"))
	})

	It("keeps only the newest capacity events", func() {
		store := &ringStore{cap: 2}
		for _, id := range []string{"1", "2", "3"} {
			store.append(newEvent(id))
		}

		replayable, err := store.EventsAfter(sse.NewEventID("0"))
		Expect(err).NotTo(HaveOccurred())
		Expect(replayable).To(HaveLen(2))
		Expect(replayable[0].ID.Get()).To(Equal("2"))
	})

	It("replays everything for a zero or unknown-format ID", func() {
		store := &ringStore{cap: 8}
		store.append(newEvent("1"))

		replayable, err := store.EventsAfter(sse.EventID{})
		Expect(err).NotTo(HaveOccurred())
		Expect(replayable).To(HaveLen(1))

		replayable, err = store.EventsAfter(sse.NewEventID("garbage"))
		Expect(err).NotTo(HaveOccurred())
		Expect(replayable).To(HaveLen(1))
	})
})

var _ = Describe("hub drop accounting", func() {
	It("counts events dropped from full subscriber buffers", func() {
		hub := NewHub(WithSubscriberBuffer(1))
		subscription := hub.broadcaster.Subscribe()
		Expect(subscription).NotTo(BeNil())

		for range 3 {
			_, err := hub.publish(EventSnapshot, Snapshot{}, &Snapshot{}, nil)
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(hub.Dropped()).To(Equal(int64(2)))

		for len(subscription) > 0 {
			<-subscription
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		Expect(hub.Shutdown(ctx)).NotTo(HaveOccurred())
	})
})
