package common

import "errors"

func ErrErrorNotFound() error {
	return errors.New("not found")
}
