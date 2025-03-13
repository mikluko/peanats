package peanats

type Submitter interface {
	Submit(func())
}

var DefaultSubmitter Submitter = submitterImpl{}

type submitterImpl struct{}

func (submitterImpl) Submit(f func()) {
	go f()
}
