package acknak

// Package provides middleware for sending back ACK, NAK and TERM acknowledgements.
//
// https://docs.nats.io/using-nats/developer/develop_jetstream/model_deep_dive

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"golang.org/x/exp/constraints"

	"github.com/mikluko/peanats"
)

type Option func(params2 *params)

// AckPolicy is a policy for sending back ACK acknowledgements.
type AckPolicy uint8

const (
	_ AckPolicy = iota
	// AckPolicyNever means that the message will never be acknowledged.
	AckPolicyNever
	// AckPolicyOnArrival means that the message will be acknowledged as soon as it arrives.
	AckPolicyOnArrival
	// AckPolicyOnSuccess means that the message will be acknowledged only if the handler returns no error.
	AckPolicyOnSuccess

	DefaultAckPolicy = AckPolicyNever
)

// NakPolicy is a policy for sending back NAK acknowledgements.
type NakPolicy uint8

const (
	_ NakPolicy = iota
	// NakPolicyNever means that the message will never be negatively acknowledged.
	NakPolicyNever

	NakPolicyOnError

	DefaultNakPolicy = NakPolicy(0)
)

// DelayPolicy calculates the delay duration for NAK acknowledgements based on the attempt number.
type DelayPolicy interface {
	// Delay calculates the delay for a given attempt number (1-based).
	Delay(attempt uint64) time.Duration
}

type params struct {
	ackPolicy      AckPolicy
	nakPolicy      NakPolicy
	nakIgnore      []error
	nakDelayPolicy DelayPolicy
	deliveryLimit  uint64
}

// MiddlewareAckPolicy sets the AckPolicy for the middleware instance.
func MiddlewareAckPolicy(policy AckPolicy) Option {
	return func(p *params) {
		p.ackPolicy = policy
	}
}

// MiddlewareNakPolicy sets the NakPolicy for the middleware instance.
func MiddlewareNakPolicy(policy NakPolicy) Option {
	return func(p *params) {
		p.nakPolicy = policy
	}
}

// MiddlewareNakIgnore sets the list of errors that should not trigger NAK.
func MiddlewareNakIgnore(errs ...error) Option {
	return func(p *params) {
		p.nakIgnore = append(p.nakIgnore, errs...)
	}
}

// MiddlewareNakDelayPolicy sets the delay policy for NAK acknowledgements.
// When set, the middleware will use NackWithDelay with delays calculated by the policy.
func MiddlewareNakDelayPolicy(policy DelayPolicy) Option {
	return func(p *params) {
		p.nakDelayPolicy = policy
	}
}

// MiddlewareDeliveryLimit sets the maximum number of times a message can be delivered.
// If the limit is reached, TERM will be sent back with an appropriate reason.
func MiddlewareDeliveryLimit[T constraints.Integer](limit T) Option {
	return func(p *params) {
		p.deliveryLimit = uint64(limit)
	}
}

const DeliveryLimitExceeded = "delivery limit exceeded"

func Middleware(opts ...Option) peanats.MsgMiddleware {
	p := params{
		ackPolicy:     DefaultAckPolicy,
		nakPolicy:     DefaultNakPolicy,
		nakIgnore:     []error{},
		deliveryLimit: 0,
	}
	for _, opt := range opts {
		opt(&p)
	}
	return func(h peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, m peanats.Msg) error {
			if p.ackPolicy == AckPolicyOnArrival {
				if err := m.(peanats.Ackable).Ack(ctx); err != nil {
					return err
				}
			}
			err := h.HandleMsg(ctx, m)
			if err != nil && p.ackPolicy != AckPolicyOnArrival {
				meta, _ := m.(peanats.Metadatable).Metadata()
				if meta != nil && p.deliveryLimit > 0 && meta.NumDelivered >= p.deliveryLimit {
					if err := m.(peanats.Ackable).TermWithReason(ctx, DeliveryLimitExceeded); err != nil {
						return err
					}
				} else if p.nakPolicy == NakPolicyOnError {
					ignore := false
					for _, e := range p.nakIgnore {
						if errors.Is(err, e) {
							ignore = true
							break
						}
					}
					if !ignore {
						var delay time.Duration
						if p.nakDelayPolicy != nil && meta != nil {
							delay = p.nakDelayPolicy.Delay(meta.NumDelivered)
						}
						if delay > 0 {
							if err := m.(peanats.Ackable).NackWithDelay(ctx, delay); err != nil {
								return err
							}
						} else {
							if err := m.(peanats.Ackable).Nak(ctx); err != nil {
								return err
							}
						}
					}
				}
			} else {
				if p.ackPolicy == AckPolicyOnSuccess {
					err = m.(peanats.Ackable).Ack(ctx)
					if err != nil {
						return err
					}
				}

			}
			return err
		})
	}
}

// applyJitter applies jitter to a delay value.
// jitter should be between 0.0 (no jitter) and 1.0 (up to 100% jitter).
// The returned delay will be in the range [delay * (1 - jitter), delay].
func applyJitter(delay time.Duration, jitter float64) time.Duration {
	if jitter <= 0 || delay <= 0 {
		return delay
	}
	if jitter > 1.0 {
		jitter = 1.0
	}
	// Random value between 0 and jitter
	jitterFactor := rand.Float64() * jitter //nolint:gosec // not cryptographic
	// Apply jitter: delay * (1 - jitterFactor)
	return time.Duration(float64(delay) * (1 - jitterFactor))
}

// ConstantDelayPolicy implements DelayPolicy with a fixed delay for all attempts.
type ConstantDelayPolicy struct {
	Duration time.Duration
	Jitter   float64 // Jitter factor (0.0 to 1.0)
}

// Delay returns the constant delay with optional jitter.
func (p *ConstantDelayPolicy) Delay(_ uint64) time.Duration {
	return applyJitter(p.Duration, p.Jitter)
}

// LinearDelayPolicy implements DelayPolicy with linear backoff.
// Delay increases linearly: base, base*2, base*3, etc., capped at max.
type LinearDelayPolicy struct {
	Base   time.Duration
	Max    time.Duration
	Jitter float64 // Jitter factor (0.0 to 1.0)
}

// Delay calculates linear backoff delay with optional jitter.
func (p *LinearDelayPolicy) Delay(attempt uint64) time.Duration {
	delay := time.Duration(attempt) * p.Base
	if p.Max > 0 && delay > p.Max {
		delay = p.Max
	}
	return applyJitter(delay, p.Jitter)
}

// ExponentialDelayPolicy implements DelayPolicy with exponential backoff.
// Delay increases exponentially: base, base*2, base*4, base*8, etc., capped at max.
type ExponentialDelayPolicy struct {
	Base   time.Duration
	Max    time.Duration
	Jitter float64 // Jitter factor (0.0 to 1.0)
}

// Delay calculates exponential backoff delay with optional jitter.
func (p *ExponentialDelayPolicy) Delay(attempt uint64) time.Duration {
	if attempt == 0 {
		return 0
	}
	// Calculate 2^(attempt-1)
	multiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(p.Base) * multiplier)
	if p.Max > 0 && delay > p.Max {
		delay = p.Max
	}
	return applyJitter(delay, p.Jitter)
}
