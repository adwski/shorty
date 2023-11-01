package errors

import "errors"

var ErrNotFound = errors.New("not found")

func Equal(err1, err2 error) bool {
	return err1.Error() == err2.Error()
}
