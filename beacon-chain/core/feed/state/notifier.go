package state

import "github.com/prysmaticlabs/prysm/v5/async/event"

// Notifier interface defines the methods of the service that provides state updates to consumers.
type Notifier interface {
	StateFeed() event.SubscriberSender
}
