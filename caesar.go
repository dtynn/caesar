package caesar

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/qiniu/log"
)

var logger = log.Std

type caesar struct {
	router  *mux.Router
	stack   *stack
	running bool
}

func New() *caesar {
	return &caesar{
		router: mux.NewRouter(),
		stack:  newStack(),
	}
}

func (this *caesar) Register(path string, f interface{}, methods ...string) {
	this.stack.addRequestHandler(path, f, methods...)
	return
}

func (this *caesar) Get(path string, f interface{}) {
	this.Register(path, f, "GET", "HEAD")
}

func (this *caesar) Post(path string, f interface{}) {
	this.Register(path, f, "POST")
}

func (this *caesar) Put(path string, f interface{}) {
	this.Register(path, f, "PUT")
}

func (this *caesar) Delete(path string, f interface{}) {
	this.Register(path, f, "DELETE")
}

func (this *caesar) Head(path string, f interface{}) {
	this.Register(path, f, "HEAD")
}

func (this *caesar) Options(path string, f interface{}) {
	this.Register(path, f, "OPTIONS")
}

func (this *caesar) Any(path string, f interface{}) {
	this.Register(path, f)
}

func (this *caesar) AddBeforeRequest(handler func(w http.ResponseWriter, r *http.Request) (int, error)) {
	this.stack.addBeforeHandler(handler)
}

func (this *caesar) SetErrorHandler(handler func(w http.ResponseWriter, r *http.Request, code int, err error)) {
	this.stack.setErrorHandler(handler)
}

func (this *caesar) AddAfterRequest(handler func(w http.ResponseWriter, r *http.Request)) {
	this.stack.addAfterHandler(handler)
}

// run
func (this *caesar) Run(addr string) error {
	if this.running {
		return fmt.Errorf("the server is already running")
	}

	this.running = true
	defer func() {
		this.running = false
	}()

	for _, h := range this.stack.requestHandlers {
		handler, err := parseHandlerFunc(h.fn, this.stack)
		if err != nil {
			return err
		}
		r := this.router.HandleFunc(h.path, handler)
		if len(h.methods) > 0 {
			r.Methods(h.methods...)
		}
	}

	this.router.HandleFunc("/{any:.*}", notFoundHandler)

	return http.ListenAndServe(addr, this.router)
}
