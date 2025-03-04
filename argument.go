package peanats

import "context"

type Argument[T any] interface {
	// Subject returns the subject of the message.
	Subject() string

	// Header returns the header of the message.
	Header() Header

	// Payload returns the typed payload of the message.
	Payload() *T
}

// Dispatcher interface defines a basic dispatcher.
type Dispatcher interface {
	Header() Header
	Error(context.Context, error)
}

type ArgumentHandler[A any] interface {
	Handle(context.Context, Dispatcher, Argument[A])
}

// ArgumentHandlerFunc is an adapter to allow the use of ordinary functions as ArgumentHandler.
type ArgumentHandlerFunc[A any] func(context.Context, Dispatcher, Argument[A])

func (f ArgumentHandlerFunc[A]) Handle(ctx context.Context, d Dispatcher, a Argument[A]) {
	f(ctx, d, a)
}
