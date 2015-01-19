package request

import (
	"bytes"
	"net/http"
	"strings"
)

type ResponseWriter struct {
	code int
	body *bytes.Buffer
	w    http.ResponseWriter
	r    *http.Request
}

func NewResponseWriter(w http.ResponseWriter, r *http.Request) *ResponseWriter {
	return &ResponseWriter{
		body: new(bytes.Buffer),
		w:    w,
		r:    r,
	}
}

func (this *ResponseWriter) WriteHeader(code int) {
	this.code = code
}

func (this *ResponseWriter) Write(b []byte) (int, error) {
	return this.body.Write(b)
}

func (this *ResponseWriter) Header() http.Header {
	return this.w.Header()
}

func (this *ResponseWriter) Output() (int, error) {
	if this.code > 0 {
		this.w.WriteHeader(this.code)
	}
	if strings.ToUpper(this.r.Method) == "HEAD" {
		return 0, nil
	}
	return this.w.Write(this.body.Bytes())
}

func (this *ResponseWriter) Reset() {
	this.code = 0
	this.body = new(bytes.Buffer)
}
