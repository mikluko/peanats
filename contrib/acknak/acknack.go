package acknak

// Package provides middleware for sending back ACK, NAK and TERM acknowledgements.
//
// https://docs.nats.io/using-nats/developer/develop_jetstream/model_deep_dive

import (
	"context"
	"errors"

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

type params struct {
	ackPolicy     AckPolicy
	nakPolicy     NakPolicy
	nakIgnore     []error
	deliveryLimit uint64
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
						if err := m.(peanats.Ackable).Nak(ctx); err != nil {
							return err
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
