package common

import "errors"

func ErrNotFound() error {
	return errors.New("not found")
}
