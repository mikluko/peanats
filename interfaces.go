package peanats

import "context"

// Unsubscriber represents something that can be unsubscribed.
type Unsubscriber interface {
	Unsubscribe() error
}

// Subscription represents a message subscription that can pull messages
// and be unsubscribed.
type Subscription interface {
	Unsubscriber
	NextMsg(ctx context.Context) (Msg, error)
}
