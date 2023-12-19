package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/adwski/shorty/internal/user"
	"github.com/golang-jwt/jwt/v5"
)

const (
	sessionCookieName = "shortySessID"

	defaultJWTCookieExpiration = 24 * time.Hour
)

type Claims struct {
	jwt.RegisteredClaims
	UID string
}

type Auth struct {
	jwtSecret string
}

func New(jwtSecret string) *Auth {
	return &Auth{jwtSecret: jwtSecret}
}

func (a *Auth) GetUser(r *http.Request) (*user.User, error) {
	sessionCookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("cannot get user from cookies: %w", err)
	}

	return a.GetUserFromJWT(sessionCookie.Value)
}

func (a *Auth) GetUserFromJWT(signedToken string) (*user.User, error) {
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

	if claims, ok := token.Claims.(*Claims); !ok {
		return nil, errors.New("token does not contain user id")
	} else {
		return user.NewWithID(claims.UID), nil
	}
}

func (a *Auth) CreateJWTCookie(u *user.User) (*http.Cookie, error) {
	token, err := a.NewToken(u)
	if err != nil {
		return nil, fmt.Errorf("cannot create auth token: %w", err)
	}
	return &http.Cookie{
		Name:  sessionCookieName,
		Value: token,
	}, nil
}

func (a *Auth) NewToken(u *user.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(defaultJWTCookieExpiration)),
		},
		UID: u.ID,
	})

	signedToken, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("cannot sign jwt token: %w", err)
	}
	return signedToken, nil
}
