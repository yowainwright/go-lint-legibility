package p

import "fmt"

func check(name string) error {
	return fmt.Errorf("cannot read %s.", name)
}
