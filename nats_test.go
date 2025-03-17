package peanats_test

import (
	"github.com/mikluko/peanats"
)

type testMsgImpl struct {
	subject string
	header  peanats.Header
	data    []byte
}

func (t testMsgImpl) Subject() string {
	return t.subject
}

func (t testMsgImpl) Header() peanats.Header {
	return t.header
}

func (t testMsgImpl) Data() []byte {
	return t.data
}
