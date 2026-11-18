// Package testsupport holds test helpers shared across hub test packages
// (external _test packages cannot share code via _test.go files).
package testsupport

import (
	"time"

	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// subscriberWaitDeadline bounds how long WaitForSubscriber polls before
// giving up (the caller's assertion then fails with a count of zero).
const subscriberWaitDeadline = 5 * time.Second

// WaitForSubscriber blocks until hub has at least one live subscriber or
// the deadline passes. SSE handlers deliver the snapshot before they
// subscribe to the feed, so a publish racing the subscribe would otherwise
// be dropped.
func WaitForSubscriber(hub *events.Hub) {
	deadline := time.Now().Add(subscriberWaitDeadline)
	for hub.SubscriberCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
}
