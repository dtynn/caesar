package caesar

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type blueprint struct {
	prefix string
	stack  *stack
}

func NewBlueprint(prefix string) (*blueprint, error) {
	if !strings.HasPrefix(prefix, "/") {
		return nil, fmt.Errorf("blueprint prefix must starts with \"/\"")
	}
	return &blueprint{
		prefix: prefix,
		stack:  newBpStack(),
	}, nil
}

func (this *blueprint) Register(path string, f interface{}, methods ...string) {
	this.stack.addRequestHandler(path, f, methods...)
}

func (this *blueprint) Get(path string, f interface{}) {
	this.Register(path, f, "GET", "HEAD")
}

func (this *blueprint) Post(path string, f interface{}) {
	this.Register(path, f, "POST")
}

func (this *blueprint) Delete(path string, f interface{}) {
	this.Register(path, f, "DELETE")
}

func (this *blueprint) Put(path string, f interface{}) {
	this.Register(path, f, "PUT")
}

func (this *blueprint) Head(path string, f interface{}) {
	this.Register(path, f, "HEAD")
}

func (this *blueprint) Options(path string, f interface{}) {
	this.Register(path, f, "OPTIONS")
}

func (this *blueprint) Any(path string, f interface{}) {
	this.Register(path, f)
}

func (this *blueprint) AddBeforeRequest(handler func(w http.ResponseWriter, r *http.Request) (int, error)) {
	this.stack.addBeforeHandler(handler)
}

func (this *blueprint) AddAfterRequest(handler func(w http.ResponseWriter, r *http.Request)) {
	this.stack.addAfterHandler(handler)
}

func (this *blueprint) SetErrorHandler(handler func(w http.ResponseWriter, r *http.Request, code int, err error)) {
	this.stack.setErrorHandler(handler)
}

func (this *blueprint) build(csr *caesar) error {
	if csr.stack == nil {
		return fmt.Errorf("app stack must not be nil")
	}

	for _, h := range this.stack.requestHandlers {
		handler, err := this.parseHandlerFunc(h.Fn, csr.stack)
		if err != nil {
			return err
		}
		logger.Debugf("blueprint handler: %s %s", h.Methods, path.Join(this.prefix, h.Path))
		r := csr.router.HandleFunc(path.Join(this.prefix, h.Path), handler)
		if len(h.Methods) > 0 {
			r.Methods(h.Methods...)
		}
	}
	return nil
}

func (this *blueprint) parseHandlerFunc(f interface{}, appStk *stack) (func(rw http.ResponseWriter, req *http.Request), error) {
	return handlerMaker(f,
		beforeHandlersMaker(appStk.beforeHandlers, this.stack.beforeHandlers),
		afterHandlersMaker(appStk.afterHandlers, this.stack.afterHandlers),
		errHanlderPicker(appStk.errorHandler, this.stack.errorHandler))
}
