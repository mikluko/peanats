package peanats

func MakePublishRespondMiddleware(msgpub MsgPublisher) Middleware {
	return func(h Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			return h.Serve(&subjectPublisher{pub, msgpub, req.Reply()}, req)
		})
	}
}
