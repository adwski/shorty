// Package filter contains filtering middleware.
// When included in chain, middleware matches incoming request
// address against configured trusted subnets and passes request
// to upstream handler only in case of successful match.
// If request address doesn't match trusted subnets, 403 response
// is sent back.
package filter

import (
	"fmt"
	"net/http"

	"github.com/adwski/shorty/internal/filter"
	"go.uber.org/zap"
)

// Middleware is a filtering middleware.
// It matches source address against trusted subnets
// and calls next handler only if there's successful match.
//
// Sources ip is taken from socket remote addr.
// Optionally source address can also be looked up in
// X-Forwarded-For and X-Real-IP headers.
type Middleware struct {
	*filter.Filter
	handler http.Handler
	block   bool
}

// Config is filtering middleware configuration.
type Config struct {
	Logger             *zap.Logger
	Subnets            string
	TrustXForwardedFor bool
	TrustXRealIP       bool
}

// New creates filtering middleware. If subnets from config
// could not be parsed, error will be returned.
func New(cfg *filter.Config) (*Middleware, error) {
	f, err := filter.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot configure filter: %w", err)
	}
	return NewFromFilter(f), nil
}

// NewFromFilter creates filter middleware using existing filter instance.
func NewFromFilter(f *filter.Filter) *Middleware {
	return &Middleware{
		Filter: f,
		block:  len(f.Subnets()) == 0,
	}
}

// HandlerFunc sets upstream middleware handler.
func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}

// ServeHTTP matches incoming request with filter
// and calls upstream handler. If there's no trusted subnets
// configured, then request is passed to next handler unconditionally.
//
// Note: For successful X-Forwarded-For match all the addresses in chain
// must be inside trusted subnets.
func (mw *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if !mw.block && mw.CheckRequestParams(
		r.RemoteAddr,
		r.Header.Get("X-Real-IP"),
		r.Header.Get("X-Forwarded-For"),
	) {
		mw.handler.ServeHTTP(rw, r)
		return
	}
	rw.WriteHeader(http.StatusForbidden)
}
