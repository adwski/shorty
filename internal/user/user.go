package user

import (
	"encoding/base64"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID  string
	new bool
}

func (u *User) IsNew() bool {
	return u.new
}

func NewWithID(id string) *User {
	return &User{ID: id}
}

func New() (*User, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate new uuid: %w", err)
	}
	return &User{
		ID:  base64.RawURLEncoding.EncodeToString(uid.Bytes()),
		new: true,
	}, nil
}

func NewFromUID(uid string) (*User, error) {
	b, err := base64.RawURLEncoding.DecodeString(uid)
	if err != nil {
		return nil, fmt.Errorf("cannot base64-decode uid: %w", err)
	}
	if _, err = uuid.FromBytes(b); err != nil {
		return nil, fmt.Errorf("cannot uuidv4-decode uid: %w", err)
	}
	return &User{ID: uid}, nil
}
