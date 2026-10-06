package p

func computed(a, b, c int) []int {
	return []int{a + b + c}
}

func functionLiteral(a, b, c int) int {
	return func() int {
		return a + b + c
	}()
}
