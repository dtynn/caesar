package caesar

import (
	"bytes"
	"net/http"
	"strings"
)

type responseWriter struct {
	code int
	body *bytes.Buffer
	w    http.ResponseWriter
	r    *http.Request
}

func newResponseWriter(w http.ResponseWriter, r *http.Request) *responseWriter {
	return &responseWriter{
		body: new(bytes.Buffer),
		w:    w,
		r:    r,
	}
}

func (this *responseWriter) WriteHeader(code int) {
	this.code = code
}

func (this *responseWriter) Write(b []byte) (int, error) {
	return this.body.Write(b)
}

func (this *responseWriter) Header() http.Header {
	return this.w.Header()
}

func (this *responseWriter) output() (int, error) {
	if this.code > 0 {
		this.w.WriteHeader(this.code)
	}
	if strings.ToUpper(this.r.Method) == "HEAD" {
		return 0, nil
	}
	return this.w.Write(this.body.Bytes())
}

func (this *responseWriter) reset() {
	this.code = 0
	this.body = new(bytes.Buffer)
}
