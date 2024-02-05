package peanats

func MakePublishSubjectMiddleware(subject string) Middleware {
	return func(h Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			return h.Serve(&publisherImpl{
				PublisherMsg: pub,
				subject:      subject,
				header:       *pub.Header(),
			}, req)
		})
	}
}
