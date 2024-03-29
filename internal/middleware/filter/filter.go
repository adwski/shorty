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
	"net/netip"
	"strings"

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
	logger       *zap.Logger
	handler      http.Handler
	subnets      []netip.Prefix
	trustXFF     bool
	trustXRealIP bool
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
func New(cfg *Config) (*Middleware, error) {
	prefixes, err := parseSliceFromString(cfg.Subnets, netip.ParsePrefix)
	if err != nil {
		return nil, err
	}
	return &Middleware{
		subnets:      prefixes,
		logger:       cfg.Logger.With(zap.String("component", "filter")),
		trustXFF:     cfg.TrustXForwardedFor,
		trustXRealIP: cfg.TrustXRealIP,
	}, nil
}

// HandlerFunc sets upstream middleware handler.
func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}

// ServeHTTP matches incoming request with trusted subnets
// and calls upstream handler. If there's no trusted subnets
// configured, then request is passed to next handler unconditionally.
//
// Parts of request is checked in order:
//   - X-Forwarded-For
//   - X-Real-IP
//   - RemoteAddr
//
// First two checks are optional (can be enabled in config), last one is always active.
//
// If in any of the stages source address matches one of the subnets, then
// this is considered success and upstream handler is called.
// If no addresses are matched then 403 is returned.
//
// Note: For successful X-Forwarded-For match all the addresses in chain
// must be inside trusted subnets.
func (mw *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if len(mw.subnets) == 0 {
		mw.handler.ServeHTTP(rw, r)
		return
	}
	if mw.trustXFF {
		if mw.processXFF(rw, r) {
			return
		}
	}
	if mw.trustXRealIP {
		if mw.processXRealIP(rw, r) {
			return
		}
	}
	if mw.processRemoteAddr(rw, r) {
		return
	}
	rw.WriteHeader(http.StatusForbidden)
}

func (mw *Middleware) processRemoteAddr(rw http.ResponseWriter, r *http.Request) bool {
	callNext, err := mw.checkIPPort(r.RemoteAddr)
	if err != nil {
		mw.logger.Debug("failed to check remote addr", zap.Error(err))
	} else if callNext {
		mw.handler.ServeHTTP(rw, r)
		return true
	}
	return false
}

func (mw *Middleware) processXFF(rw http.ResponseWriter, r *http.Request) bool {
	addrs, errXFF := parseSliceFromString(r.Header.Get("X-Forwarded-For"), netip.ParseAddr)
	if errXFF != nil {
		mw.logger.Debug("failed to get IPs from X-Forwarded-For", zap.Error(errXFF))
		return false
	}
	if len(addrs) == 0 {
		return false
	}
Loop:
	for i := range addrs {
		for j := range mw.subnets {
			if mw.subnets[j].Contains(addrs[i]) {
				continue Loop
			}
		}
		return false
	}
	mw.handler.ServeHTTP(rw, r)
	return true
}

func (mw *Middleware) processXRealIP(rw http.ResponseWriter, r *http.Request) bool {
	callNext, err := mw.checkIP(r.Header.Get("X-Real-IP"))
	if err != nil {
		mw.logger.Debug("failed to check X-Real-IP", zap.Error(err))
	} else if callNext {
		mw.handler.ServeHTTP(rw, r)
		return true
	}
	return false
}

func (mw *Middleware) checkIP(ip string) (bool, error) {
	if ip == "" {
		return false, nil
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false, fmt.Errorf("cannot parse ip address string: %w", err)
	}
	for i := range mw.subnets {
		if mw.subnets[i].Contains(addr) {
			return true, nil
		}
	}
	return false, nil
}

func (mw *Middleware) checkIPPort(ip string) (bool, error) {
	if ip == "" {
		return false, nil
	}
	addr, err := netip.ParseAddrPort(ip)
	if err != nil {
		return false, fmt.Errorf("cannot parse ip:port address string: %w", err)
	}
	for i := range mw.subnets {
		if mw.subnets[i].Contains(addr.Addr()) {
			return true, nil
		}
	}
	return false, nil
}

// parseSliceFromString parses comma separated list of arbitrary types.
// Input string is split by comma and then each resulting string
// is passed to parse function which instantiates type from string.
// Resulting slice of types is returned.
//
// String values should not contain leading spaces (they will be trimmed).
//
// In the scope of this package parseSliceFromString is used to parse IP prefixes and addresses.
// It also complies to X-Forwarded-For spec (hence specific spaces handling):
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For.
//
// If any of entities in string cannot be parsed, function returns nil slice and error.
func parseSliceFromString[T any](s string, parse func(string) (T, error)) ([]T, error) {
	if s == "" {
		return nil, nil
	}
	entitiesS := strings.Split(s, ",")
	entities := make([]T, 0, len(entitiesS))
	for i := range entitiesS {
		entity, err := parse(strings.TrimLeft(entitiesS[i], " "))
		if err != nil {
			return nil, fmt.Errorf("cannot parse entity: %w", err)
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
