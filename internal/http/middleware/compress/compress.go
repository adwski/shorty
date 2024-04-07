// Package compress implements compression middleware.
//
// Supported algorithms are gzip and deflate.
// Only application/json and text/plain will be compressed if client
// set appropriate Accept-Encoding value.
package compress

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
)

const (
	typeGZIP    = "gzip"
	typeDeflate = "deflate"
)

// Middleware is compress middleware.
type Middleware struct {
	handler http.Handler
}

// New creates new compress middleware.
func New() *Middleware {
	return &Middleware{}
}

type rwWrapper struct {
	http.ResponseWriter
	cw io.WriteCloser
	ct string
}

func newResponseWrapper(w http.ResponseWriter, enc string) *rwWrapper {
	rw := &rwWrapper{ResponseWriter: w}
	if strings.Contains(enc, typeGZIP) {
		rw.ct = typeGZIP
	} else if strings.Contains(enc, typeDeflate) {
		rw.ct = typeDeflate
	}
	return rw
}

func (rw *rwWrapper) close() {
	if rw.cw != nil {
		_ = rw.cw.Close()
	}
}

// Write writes body either to compression writer or to response body directly depending on
// whether compression writer is nil or not.
func (rw *rwWrapper) Write(b []byte) (n int, err error) {
	if rw.cw != nil {
		n, err = rw.cw.Write(b)
	} else {
		n, err = rw.ResponseWriter.Write(b)
	}
	return
}

func (rw *rwWrapper) needCompression() bool {
	switch rw.Header().Get("Content-Type") {
	case "application/json",
		"text/plain":
		return true
	}
	return false
}

// WriteHeader writes status code and Content-Encoding if compression writer is set and response needs compression.
func (rw *rwWrapper) WriteHeader(statusCode int) {
	if rw.ct != "" && rw.needCompression() {
		rw.ResponseWriter.Header().Set("Content-Encoding", rw.ct)
		switch rw.ct {
		case typeGZIP:
			rw.cw = gzip.NewWriter(rw.ResponseWriter)
		case typeDeflate:
			rw.cw = zlib.NewWriter(rw.ResponseWriter)
		}
	}
	rw.ResponseWriter.WriteHeader(statusCode)
}

// HandlerFunc sets upstream middleware handler.
func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}

// ServeHTTP wraps http response writer to compression writer for upstream handlers.
func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := newResponseWrapper(w, r.Header.Get("Accept-Encoding"))
	defer func() {
		rw.close()
		rw = nil
	}()
	mw.handler.ServeHTTP(rw, r)
}
