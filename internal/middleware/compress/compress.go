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

type Middleware struct {
	handler http.Handler
}

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

func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
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
