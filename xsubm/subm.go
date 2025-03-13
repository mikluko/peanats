package xsubm

type Submitter interface {
	Submit(func())
}

type JustGoSubmitter struct{}

func (JustGoSubmitter) Submit(f func()) {
	go f()
}
