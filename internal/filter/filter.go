// Package filter contains ip filter.
// It provides functions to match <ip>, <ip:port> or <ip, ip, ..., ip> strings against
// configured trusted subnets.
package filter

import (
	"fmt"
	"net/netip"
	"strings"

	"go.uber.org/zap"
)

// Filter is an ip filter.
//
// CheckIPPort(), CheckIP(), CheckXFF() can be used for matching.
type Filter struct {
	logger       *zap.Logger
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

// New creates filter. If subnets from config
// could not be parsed, error will be returned.
func New(cfg *Config) (*Filter, error) {
	prefixes, err := parseSliceFromString(cfg.Subnets, netip.ParsePrefix)
	if err != nil {
		return nil, err
	}
	return &Filter{
		subnets:      prefixes,
		logger:       cfg.Logger.With(zap.String("component", "filter")),
		trustXFF:     cfg.TrustXForwardedFor,
		trustXRealIP: cfg.TrustXRealIP,
	}, nil
}

// Subnets returns parsed trusted subnets.
func (f *Filter) Subnets() []netip.Prefix {
	return f.subnets
}

// CheckRequestParams matches incoming request parameters with trusted subnets
// and returns true if at least one of three parameters matches.
// If there's no trusted subnets configured, then true is always returned.
//
// Parts of request is checked in order:
//   - RemoteAddr
//   - X-Real-IP
//   - X-Forwarded-For
//
// Last two checks are optional (can be enabled in config), first one is always active.
//
// Note: For successful X-Forwarded-For match all the addresses in chain
// must be inside trusted subnets.
func (f *Filter) CheckRequestParams(remoteAddr, xRealIP, xForwardedFor string) bool {
	if f.CheckIPPort(remoteAddr) {
		return true
	}
	if f.trustXRealIP {
		if f.CheckIP(xRealIP) {
			return true
		}
	}
	if f.trustXFF {
		if f.CheckIPList(xForwardedFor) {
			return true
		}
	}
	return false
}

// CheckIPPort checks whether provided address:port string
// matches against configured trusted subnets.
func (f *Filter) CheckIPPort(ipPort string) bool {
	ok, err := f.checkIPPort(ipPort)
	if err != nil {
		f.logger.Debug("failed to check remote addr", zap.Error(err))
	} else if ok {
		return true
	}
	return false
}

// CheckIPList checks whether all IP addresses in <ip, ip, ..., ip> string
// match against configured trusted subnets.
func (f *Filter) CheckIPList(ipList string) bool {
	addrs, errXFF := parseSliceFromString(ipList, netip.ParseAddr)
	if errXFF != nil {
		f.logger.Debug("failed to get IPs from X-Forwarded-For", zap.Error(errXFF))
		return false
	}
	if len(addrs) == 0 {
		return false
	}
Loop:
	for i := range addrs {
		for j := range f.subnets {
			if f.subnets[j].Contains(addrs[i]) {
				continue Loop
			}
		}
		return false
	}
	return true
}

// CheckIP checks whether provided IP address string
// matches against configured trusted subnets.
func (f *Filter) CheckIP(ip string) bool {
	callNext, err := f.checkIP(ip)
	if err != nil {
		f.logger.Debug("failed to check X-Real-IP", zap.Error(err))
	} else if callNext {
		return true
	}
	return false
}

func (f *Filter) checkIP(ip string) (bool, error) {
	if ip == "" {
		return false, nil
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false, fmt.Errorf("cannot parse ip address string: %w", err)
	}
	for i := range f.subnets {
		if f.subnets[i].Contains(addr) {
			return true, nil
		}
	}
	return false, nil
}

func (f *Filter) checkIPPort(ip string) (bool, error) {
	if ip == "" {
		return false, nil
	}
	addr, err := netip.ParseAddrPort(ip)
	if err != nil {
		return false, fmt.Errorf("cannot parse ip:port address string: %w", err)
	}
	for i := range f.subnets {
		if f.subnets[i].Contains(addr.Addr()) {
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
