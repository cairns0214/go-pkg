package log

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}
