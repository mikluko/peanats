package jetconsumer

import (
	"context"
	"errors"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type JetstreamFetcher interface {
	Fetch(batch int, opts ...jetstream.FetchOpt) (jetstream.MessageBatch, error)
}

type BatchConsumer struct {
	BaseContext context.Context
	Fetcher     JetstreamFetcher
	Handler     BatchHandler
	Executor    func(func())

	BatchSize int
	Wait      time.Duration

	cancel context.CancelFunc
	err    error
}

func (c *BatchConsumer) Start() (err error) {
	if c.Fetcher == nil {
		panic("jetstream consumer is not set")
	}
	if c.Handler == nil {
		panic("handler is not set")
	}
	if c.BatchSize == 0 {
		panic("batch size is not set")
	}

	if c.BaseContext == nil {
		c.BaseContext = context.Background()
	}
	c.BaseContext, c.cancel = context.WithCancel(c.BaseContext)

	if c.Executor == nil {
		c.Executor = func(f func()) { f() }
	}

	go c.loop()

	return nil
}

func (c *BatchConsumer) Stop() {
	c.cancel()
}

func (c *BatchConsumer) Error() error {
	return c.err
}

func (c *BatchConsumer) loop() {
	opts := make([]jetstream.FetchOpt, 0)
	if c.Wait > 0 {
		opts = append(opts, jetstream.FetchMaxWait(c.Wait))
	}
	for {
		select {
		case <-c.BaseContext.Done():
			return
		default:
			batch, err := c.Fetcher.Fetch(c.BatchSize, opts...)
			if errors.Is(err, jetstream.ErrNoMessages) {
				continue
			} else if err != nil {
				c.err = err
				return
			}
			c.Executor(func() {
				err = c.Handler.Serve(c.BaseContext, batch)
				if err != nil {
					panic(err)
				}
			})
		}
	}
}
