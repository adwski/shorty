// Package user describes user object and all its internal info.
package user

import (
	"encoding/base64"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// User represents single user.
type User struct {
	ID  string
	new bool
}

// IsNew returns whether user was created (generated) during this request (is new) or
// it was retrieved from jwt cookie.
func (u *User) IsNew() bool {
	return u.new
}

// NewWithID create user object with particular UD. It should be used to instantiate user
// after successful cookie parse.
func NewWithID(id string) *User {
	return &User{ID: id}
}

// New create new user object with generated ID.
func New() (*User, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate new uuid: %w", err)
	}
	return &User{
		ID:  base64.RawURLEncoding.EncodeToString(u.Bytes()),
		new: true,
	}, nil
}

// NewFromUserID is same as NewWithID but also validates that userID is correct base64 encoded UUIDv4 bytes.
func NewFromUserID(userID string) (*User, error) {
	b, err := base64.RawURLEncoding.DecodeString(userID)
	if err != nil {
		return nil, fmt.Errorf("cannot base64-decode user id: %w", err)
	}
	if _, err = uuid.FromBytes(b); err != nil {
		return nil, fmt.Errorf("cannot uuidv4-decode user id: %w", err)
	}
	return NewWithID(userID), nil
}
