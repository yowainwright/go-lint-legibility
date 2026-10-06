package p

func check(a, b, c bool) {
	outer(inner(
		a && b && c,
	))
	consume(func() bool {
		return a && b && c
	})
}
