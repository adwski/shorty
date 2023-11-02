package errors

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrGiveUP        = errors.New("given up creating redirect")
)

func Equal(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true // not sure
	} else if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}
