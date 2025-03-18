package peanats

type Submitter interface {
	Submit(func())
}

type SubmitterFunc func(func())

func (f SubmitterFunc) Submit(g func()) {
	f(g)
}

var DefaultSubmitter Submitter = submitterImpl{}

type submitterImpl struct{}

func (submitterImpl) Submit(f func()) {
	go f()
}
