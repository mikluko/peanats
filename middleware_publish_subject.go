package peanats

func MakePublishSubjectMiddleware(msgpub MsgPublisher, subject string) Middleware {
	return func(h Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			return h.Serve(&subjectPublisher{pub, msgpub, subject}, req)
		})
	}
}
