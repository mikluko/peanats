package pond

import (
	"github.com/alitto/pond/v2"

	"github.com/mikluko/peanats"
)

func Submitter(maxConcurrency int, opts ...pond.Option) peanats.Submitter {
	pool := pond.NewPool(maxConcurrency, opts...)
	return &submitterImpl{pool}
}

func SubmitterPool(pool pond.Pool) peanats.Submitter {
	return &submitterImpl{pool}
}

type submitterImpl struct {
	pool pond.Pool
}

func (p submitterImpl) Submit(f func()) {
	err := p.pool.Go(f)
	if err != nil {
		panic(err)
	}
}
