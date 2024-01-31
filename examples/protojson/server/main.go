package main

import (
	"os"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/examples/protojson/api"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	hnd := peanats.TypedHandler[api.Argument, api.Result](peanats.TypedHandlerFunc[api.Argument, api.Result](handle))
	srv := peanats.Server{
		ListenSubjects: []string{"peanuts.protojson.requests"},
		Conn:           nc,
		Handler: peanats.ChainMiddleware(
			peanats.Typed(&peanats.ProtojsonCodec{}, hnd),
			peanats.MakePublishSubjectMiddleware(nc, "peanuts.protojson.results"),
			peanats.MakeAccessLogMiddleware(os.Stdout),
		),
	}

	err = srv.Start()
	if err != nil {
		panic(err)
	}
	err = srv.Wait()
	if err != nil {
		panic(err)
	}
}

func handle(pub peanats.TypedPublisher[api.Result], req peanats.TypedRequest[api.Argument]) error {
	return pub.Publish(&api.Result{Res: req.Payload().Arg})
}
