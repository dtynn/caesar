package caesar

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type Blueprint struct {
	mutex sync.Mutex
	built bool

	prefix string
	stack  *stack
}

func NewBlueprint(prefix string) (*Blueprint, error) {
	if !strings.HasPrefix(prefix, "/") {
		return nil, fmt.Errorf("blueprint prefix must starts with \"/\"")
	}
	return &Blueprint{
		prefix: prefix,
		stack:  newBpStack(),
	}, nil
}

func (this *Blueprint) Register(path string, f interface{}, methods ...string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.addRequestHandler(path, f, methods...)
}

func (this *Blueprint) Get(path string, f interface{}) {
	this.Register(path, f, "GET", "HEAD")
}

func (this *Blueprint) Post(path string, f interface{}) {
	this.Register(path, f, "POST")
}

func (this *Blueprint) Delete(path string, f interface{}) {
	this.Register(path, f, "DELETE")
}

func (this *Blueprint) Put(path string, f interface{}) {
	this.Register(path, f, "PUT")
}

func (this *Blueprint) Head(path string, f interface{}) {
	this.Register(path, f, "HEAD")
}

func (this *Blueprint) Options(path string, f interface{}) {
	this.Register(path, f, "OPTIONS")
}

func (this *Blueprint) Any(path string, f interface{}) {
	this.Register(path, f)
}

func (this *Blueprint) AddBeforeRequest(handler func(w http.ResponseWriter, r *http.Request) (int, error)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.addBeforeHandler(handler)
}

func (this *Blueprint) AddAfterRequest(handler func(w http.ResponseWriter, r *http.Request)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.addAfterHandler(handler)
}

func (this *Blueprint) SetErrorHandler(handler func(w http.ResponseWriter, r *http.Request, code int, err error)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.setErrorHandler(handler)
}

func (this *Blueprint) SetNotFoundHandler(handler func(w http.ResponseWriter, r *http.Request)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.setNotFoundHandler(handler)
}

func (this *Blueprint) build(csr *Caesar) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.built {
		return fmt.Errorf("this blueprint has been built")
	}

	if csr.stack == nil {
		return fmt.Errorf("app stack must not be nil")
	}

	var err error

	prefix := this.prefix
	if p := csr.Config().Prefix; p != "" {
		prefix, err = makeRequestURI(p, prefix)
		if err != nil {
			return err
		}
	}

	for _, h := range this.stack.requestHandlers {
		handler, err := this.parseHandlerFunc(h.Fn, csr.stack)
		if err != nil {
			return err
		}
		path, err := makeRequestURI(prefix, h.Path)
		if err != nil {
			return err
		}

		logger.Debugf("blueprint handler: %s %s", h.Methods, path)
		r := csr.router.HandleFunc(path, handler)
		if len(h.Methods) > 0 {
			r.Methods(h.Methods...)
		}
	}

	if this.stack.notFoundHandler != nil {
		bpAnyHandler := this.stack.notFoundHandler

		bpAnyPath, err := makeRequestURI(this.prefix, anyPath)
		if err != nil {
			return err
		}

		csr.router.HandleFunc(bpAnyPath, bpAnyHandler)
	}

	this.built = true

	return nil
}

func (this *Blueprint) parseHandlerFunc(f interface{}, appStk *stack) (func(rw http.ResponseWriter, req *http.Request), error) {
	return handlerMaker(f,
		beforeHandlersMaker(appStk.beforeHandlers, this.stack.beforeHandlers),
		afterHandlersMaker(appStk.afterHandlers, this.stack.afterHandlers),
		errHanlderPicker(appStk.errorHandler, this.stack.errorHandler))
}
