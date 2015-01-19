package caesar

import (
	"fmt"
	"net/http"

	"github.com/dtynn/caesar/request"
	"github.com/gorilla/mux"
	"github.com/qiniu/log"
)

var logger = log.Std

type caesar struct {
	blueprints []*blueprint
	router     *mux.Router
	stack      *stack
	running    bool
}

func New() *caesar {
	return &caesar{
		blueprints: []*blueprint{},
		router:     mux.NewRouter(),
		stack:      newAppStack(),
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

func (this *caesar) RegisterBlueprint(bp *blueprint) {
	if bp != nil {
		this.blueprints = append(this.blueprints, bp)
	}
}

func (this *caesar) AddBeforeRequest(handler func(w http.ResponseWriter, r *http.Request) (int, error)) {
	this.stack.addBeforeHandler(handler)
}

func (this *caesar) AddAfterRequest(handler func(w http.ResponseWriter, r *http.Request)) {
	this.stack.addAfterHandler(handler)
}

func (this *caesar) SetErrorHandler(handler func(w http.ResponseWriter, r *http.Request, code int, err error)) {
	this.stack.setErrorHandler(handler)
}

func (this *caesar) parseHandlerFunc(f interface{}) (func(rw http.ResponseWriter, req *http.Request), error) {
	return handlerMaker(f, this.stack.beforeHandlers, this.stack.afterHandlers, this.stack.errorHandler)
}

func (this *caesar) build() error {
	for _, h := range this.stack.requestHandlers {
		handler, err := this.parseHandlerFunc(h.Fn)
		if err != nil {
			return err
		}
		logger.Debugf("app handler: %s %s", h.Methods, h.Path)
		r := this.router.HandleFunc(h.Path, handler)
		if len(h.Methods) > 0 {
			r.Methods(h.Methods...)
		}
	}
	return nil
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

	// build blueprint router
	for _, bp := range this.blueprints {
		if err := bp.build(this); err != nil {
			return err
		}
	}

	// build app route
	if err := this.build(); err != nil {
		return err
	}

	this.router.HandleFunc("/{any:.*}", request.DefaultNotFoundHandler)

	return http.ListenAndServe(addr, this.router)
}
