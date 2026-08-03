package p

var _ bool

func check(is bool, has bool, _ bool) bool {
	return is || has
}

type options struct {
	_ bool
}
