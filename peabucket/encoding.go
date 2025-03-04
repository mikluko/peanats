package peabucket

import (
	"bytes"
	"mime/multipart"
	"net/textproto"

	"github.com/mikluko/peanats"
)

const boundary = "--"

func encode[T any](h textproto.MIMEHeader, v *T) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	if err := w.SetBoundary(boundary); err != nil {
		panic(err)
	}
	c := peanats.ChooseCodec(h)
	c.SetHeader(h)
	p, err := c.Encode(v)
	if err != nil {
		return nil, err
	}
	q, _ := w.CreatePart(h)
	_, _ = q.Write(p)
	return buf.Bytes(), nil
}

func decode[T any](b []byte) (_ textproto.MIMEHeader, _ *T, err error) {
	r := multipart.NewReader(bytes.NewReader(b), boundary)
	p, err := r.NextPart()
	if err != nil {
		return nil, nil, err
	}
	h := p.Header
	c := peanats.ChooseCodec(h)
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(p)
	v := new(T)
	err = c.Decode(buf.Bytes(), v)
	if err != nil {
		return nil, nil, err
	}
	return h, v, nil
}
