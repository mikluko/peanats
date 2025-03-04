package xargpool

import (
	"context"
	"math"

	"github.com/jackc/puddle/v2"
)

type Pool[T any] interface {
	Acquire(ctx context.Context) PoolValue[T]
}

type PoolValue[T any] interface {
	Value() *T
	Release()
}

func New[T any]() Pool[T] {
	pool, err := puddle.NewPool(&puddle.Config[*T]{
		MaxSize: int32(math.MaxInt32),
		Constructor: func(_ context.Context) (*T, error) {
			return new(T), nil
		},
		Destructor: func(v *T) {
			if a, ok := any(v).(interface{ Reset() }); ok {
				a.Reset()
			}
		},
	})
	if err != nil {
		panic(err)
	}
	return &argumentPool[T]{pool}
}

type argumentPool[T any] struct {
	pool *puddle.Pool[*T]
}

func (p *argumentPool[T]) Acquire(ctx context.Context) PoolValue[T] {
	x, err := p.pool.Acquire(ctx)
	if err != nil {
		panic(err)
	}
	return PoolValue[T](x)
}
