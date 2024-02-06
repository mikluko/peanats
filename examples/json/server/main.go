package main

import (
	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/examples/json/api"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	hnd := peanats.TypedHandler[api.Argument, api.Result](peanats.TypedHandlerFunc[api.Argument, api.Result](handle))
	srv := peanats.Server{
		ListenSubjects: []string{"peanuts.json.requests"},
		Conn:           peanats.NATS(nc),
		Handler: peanats.ChainMiddleware(
			peanats.Typed(&peanats.JsonCodec{}, hnd),
			peanats.MakePublishSubjectMiddleware("peanuts.json.results"),
			peanats.MakeAccessLogMiddleware(),
		),
	}

	err = srv.Start()
	if err != nil {
		panic(err)
	}

	srv.Wait()
}

func handle(pub peanats.TypedPublisher[api.Result], req peanats.TypedRequest[api.Argument]) error {
	return pub.Publish(&api.Result{Res: req.Payload().Arg})
}
