package xtestutil

func Must[T any](arg T, err error) T {
	if err != nil {
		panic(err)
	}
	return arg
}
