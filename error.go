package peanats

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Error struct {
	Code    int
	Message string
	Cause   error
}

func (e Error) Error() string {
	var bits []string
	if e.Code != 0 {
		bits = append(bits, fmt.Sprintf("code=%d", e.Code))
	}
	if e.Message == "" {
		bits = append(bits, "micronats error")
	} else {
		bits = append(bits, e.Message)
	}
	if e.Cause != nil {
		bits = append(bits)
	}
	return strings.Join(bits, ": ")
}

func (e Error) Unwrap() error {
	return e.Cause
}

const (
	HeaderErrorCode    = "error-code"
	HeaderErrorMessage = "error-message"
)

func prepareError(pub Publisher, err error) {
	owl := new(Error)
	if errors.As(err, owl) {
		pub.Header().Set(HeaderErrorCode, strconv.Itoa(owl.Code))
		pub.Header().Set(HeaderErrorMessage, owl.Error())
	} else {
		pub.Header().Set(HeaderErrorCode, strconv.Itoa(http.StatusInternalServerError))
		pub.Header().Set(HeaderErrorMessage, err.Error())
	}
}

func AckError(ack AckPublisher, err error) error {
	prepareError(ack, err)
	return ack.Ack(nil)
}

func PublishError(pub Publisher, err error) error {
	prepareError(pub, err)
	return pub.Publish(nil)
}
