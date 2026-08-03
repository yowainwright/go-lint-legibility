package p

import "errors"

func check() error {
	return errors.New("Invalid user")
}
