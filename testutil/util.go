package testutil

func must[T any](arg T, err error) T {
	if err != nil {
		panic(err)
	}
	return arg
}
