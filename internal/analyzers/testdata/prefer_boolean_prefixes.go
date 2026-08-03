package p

var active bool
var visible bool

func load(enabled bool, ready bool) (valid bool, done bool) {
	return enabled, ready
}
