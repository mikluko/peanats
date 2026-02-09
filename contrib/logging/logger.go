package logging

import "context"

// Logger is a simple logging interface used by the logging middlewares.
type Logger interface {
	Log(ctx context.Context, message string, args ...any)
}
