package main

import "testing"

func TestHasVersionFlag(t *testing.T) {
	cases := map[string]struct {
		args []string
		want bool
	}{
		"version only":       {[]string{"--version"}, true},
		"single dash only":   {[]string{"-version"}, true},
		"after package":      {[]string{"./...", "--version"}, false},
		"with package after": {[]string{"--version", "./..."}, false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := hasVersionFlag(c.args); got != c.want {
				t.Fatalf("hasVersionFlag(%v) = %v, want %v", c.args, got, c.want)
			}
		})
	}
}
