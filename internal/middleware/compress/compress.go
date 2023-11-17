package compress

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
)

type Middleware struct {
	handler http.Handler
}

func New() *Middleware {
	return &Middleware{}
}

type rwWrapper struct {
	w  http.ResponseWriter
	cw io.WriteCloser
	ct string
}

func newResponseWrapper(w http.ResponseWriter, enc string) *rwWrapper {
	rw := &rwWrapper{w: w}
	if strings.Contains(enc, "gzip") {
		rw.ct = "gzip"
	} else if strings.Contains(enc, "deflate") {
		rw.ct = "deflate"
	}
	return rw
}

func (rw *rwWrapper) close() {
	if rw.cw != nil {
		_ = rw.cw.Close()
	}
}

func (rw *rwWrapper) Write(b []byte) (n int, err error) {
	if rw.cw != nil && rw.needCompression() {
		n, err = rw.cw.Write(b)
	} else {
		n, err = rw.w.Write(b)
	}
	return
}

func (rw *rwWrapper) needCompression() bool {
	switch rw.w.Header().Get("Content-Type") {
	case "application/json",
		"text/plain":
		return true
	}
	return false
}

func (rw *rwWrapper) WriteHeader(statusCode int) {
	if rw.ct != "" && rw.needCompression() {
		rw.w.Header().Set("Content-Encoding", rw.ct)
		switch rw.ct {
		case "gzip":
			rw.cw = gzip.NewWriter(rw.w)
		case "deflate":
			rw.cw = zlib.NewWriter(rw.w)
		}
	}
	rw.w.WriteHeader(statusCode)
}

func (rw *rwWrapper) Header() http.Header {
	return rw.w.Header()
}

func (mw *Middleware) Chain(h http.Handler) *Middleware {
	mw.handler = h
	return mw
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := newResponseWrapper(w, r.Header.Get("Accept-Encoding"))
	defer func() {
		rw.close()
		rw = nil
	}()
	mw.handler.ServeHTTP(rw, r)
}
