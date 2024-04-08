// Package auth implements user authenticator.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/adwski/shorty/internal/user"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultJWTCookieExpiration = 24 * time.Hour
)

// Claims is jwt token claims.
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id,omitempty"`
}

// Auth is authenticator component providing hi level user operations.
type Auth struct {
	jwtSecret string
}

// New creates authenticator.
func New(jwtSecret string) *Auth {
	return &Auth{jwtSecret: jwtSecret}
}

// CreateOrParseUserFromJWTString parses user from jwt string.
// If user is parsed successfully and jwt is not expired, parsed user is returned
// and returned cookie will be nil. If user could not be parsed, new user will be created
// and corresponding cookie value will be generated. In this case both new user and cookie are returned.
func (a *Auth) CreateOrParseUserFromJWTString(usr string) (*user.User, string, error) {
	u, err := a.getUserFromJWT(usr)
	if err == nil {
		return u, "", nil
	}
	// Missing or invalid session cookie
	// Generate a new user
	return a.CreateUserAndToken()
}

// createJWT creates new jwt token for specified user.
func (a *Auth) createJWT(u *user.User) (string, error) {
	token, err := a.newToken(u)
	if err != nil {
		return "", fmt.Errorf("cannot create auth token: %w", err)
	}
	return token, nil
}

// CreateUserAndToken creates new user and corresponding jwt token.
func (a *Auth) CreateUserAndToken() (*user.User, string, error) {
	// Create user with unique ID
	u, err := user.New()
	if err != nil {
		return nil, "", fmt.Errorf("cannot create new user: %w", err)
	}

	// Set Cookie for new user
	sessionCookie, errS := a.createJWT(u)
	if errS != nil {
		return nil, "", fmt.Errorf("cannot create session cookie for user: %w", errS)
	}
	return u, sessionCookie, nil
}

func (a *Auth) getUserFromJWT(signedToken string) (*user.User, error) {
	token, err := jwt.ParseWithClaims(signedToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot parse token from session cookie: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}
	claims, ok := token.Claims.(*Claims)
	switch {
	case !ok:
		return nil, errors.New("token does not have claims")
	case claims.ExpiresAt == nil:
		return nil, errors.New("expiration claim is missing")
	case claims.ExpiresAt.Before(time.Now()):
		return nil, errors.New("token expired")
	case claims.UserID == "":
		return nil, errors.New("user id is empty")
	default:
		return user.NewWithID(claims.UserID), nil
	}
}

func (a *Auth) newToken(u *user.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(defaultJWTCookieExpiration)),
		},
		UserID: u.ID,
	})

	signedToken, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("cannot sign jwt token: %w", err)
	}
	return signedToken, nil
}
