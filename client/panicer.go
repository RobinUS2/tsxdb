package client

func panicOnErrorClose(fn func() error) {
	err := fn()
	if err != nil {
		panic(err)
	}
}
