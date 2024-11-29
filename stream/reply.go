package stream

import (
	crand "crypto/rand"
	"encoding/base32"
	"fmt"
	mrand "math/rand"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"
)

type ReplySubjecter interface {
	ReplySubject() string
}

type ReplySubjecterFunc func() string

func (f ReplySubjecterFunc) ReplySubject() string {
	return f()
}

// ReplySubjectInbox returns reply subject generator based on nats.NewInbox
func ReplySubjectInbox() ReplySubjecter {
	return ReplySubjecterFunc(func() string {
		return nats.NewInbox()
	})
}

// ReplySubjectNUID returns fast stream reply subject generator
func ReplySubjectNUID() ReplySubjecter {
	return ReplySubjecterFunc(func() string {
		return fmt.Sprintf("_STREAM.%s", nuid.Next())
	})
}

// ReplySubjectRand returns reply subject generator based on pseudo-randomness
func ReplySubjectRand() ReplySubjecter {
	return ReplySubjecterFunc(func() string {
		p := [35]byte{} // base32 encodes 35 bytes into 56 characters without padding
		_, _ = mrand.Read(p[:])
		s := base32.StdEncoding.EncodeToString(p[:])
		return fmt.Sprintf("_STREAM.%s", s)
	})
}

// ReplySubjectCryptoRand returns reply subject generator based on cryptographic-grade randomness
func ReplySubjectCryptoRand() ReplySubjecter {
	return ReplySubjecterFunc(func() string {
		p := [35]byte{} // base32 encodes 35 bytes into 56 characters without padding
		_, _ = crand.Read(p[:])
		s := base32.StdEncoding.EncodeToString(p[:])
		return fmt.Sprintf("_STREAM.%s", s)
	})
}
