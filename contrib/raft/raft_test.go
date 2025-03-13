//go:build acc

package raft_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/nats-io/nats.go"
	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/sync/semaphore"

	"github.com/mikluko/peanats/contrib/raft"
)

// Example demonstrates how to use the raft package. Requires running NATS server on localhost:4222.
//
// Running with below command will produce verbose output on each run, disregarding the cache.
//
//	go test -count=1 -tags=acc -v ./raft
//
// To also watch Raft negotiation on NATS run:
//
//	nats subscribe 'graft.>'
//
// Happy rafting.
func Example() {
	slog.SetDefault(slog.New(slogmulti.Fanout(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}
				return a
			},
		}),
	)))

	sem := semaphore.NewWeighted(3)
	for i := 0; i < 5; i++ {
		_ = sem.Acquire(context.Background(), 1)
		go func() {
			defer sem.Release(1)
			state, err := worker(i)
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				slog.Error("errored", slog.Int("worker", i), slog.Any("err", err))
			} else {
				slog.Info("completed", slog.Int("worker", i), slog.String("state", state.String()))
			}
		}()
		time.Sleep(time.Second * 4)
	}

	_ = sem.Acquire(context.Background(), 3)
	// Output:
	// level=INFO msg=completed worker=0 state=Leader
	// level=INFO msg=completed worker=1 state=Leader
	// level=INFO msg=completed worker=2 state=Leader
	// level=INFO msg=completed worker=3 state=Leader
	// level=INFO msg=completed worker=4 state=Candidate
}

type stateChangeHandler struct {
	log *slog.Logger
}

func (h *stateChangeHandler) HandleStateChange(ctx context.Context, from raft.State, to raft.State) error {
	h.log.Log(ctx, slog.LevelDebug-1, "state change", "from", from, "to", to)
	return nil
}

func worker(idx int) (raft.State, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log := slog.With("worker", idx)
	log.Debug("joining")
	defer log.Debug("leaving")

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	r, err := raft.New(nc,
		raft.WithSize(3),
		raft.WithContext(ctx),
		raft.WithLog(fmt.Sprintf(path.Join(os.TempDir(), "raft.%d.log"), idx)),
		raft.WithStateChangeHandler(&stateChangeHandler{log: log}),
	)
	if err != nil {
		return raft.RaftStateFollower, err
	}

	var state raft.State

	state = r.State()
	log.Debug("reached", slog.String("state", state.String()))

	err = r.Wait(ctx, raft.RaftStateCandidate)
	if err != nil {
		return state, err
	}

	state = r.State()
	log.Debug("reached", slog.String("state", state.String()))

	err = r.Wait(ctx, raft.RaftStateLeader)
	if err != nil {
		return state, err
	}

	state = r.State()
	log.Debug("reached", slog.String("state", state.String()))

	return state, nil
}
