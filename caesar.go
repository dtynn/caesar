package caesar

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/dtynn/caesar/request"
	"github.com/gorilla/mux"
	"github.com/qiniu/log"
)

var (
	logger         = log.Std
	makeRequestURI = request.MakeRequestURI
	anyPath        = "/{any:.*}"
)

type caesar struct {
	mutex   sync.Mutex
	running bool
	debug   bool

	blueprints []*blueprint
	router     *mux.Router
	stack      *stack
}

func New() *caesar {
	return &caesar{
		blueprints: []*blueprint{},
		router:     mux.NewRouter(),
		stack:      newAppStack(),
	}
}

func (this *caesar) Register(path string, f interface{}, methods ...string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

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
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if bp != nil {
		this.blueprints = append(this.blueprints, bp)
	}
}

func (this *caesar) AddBeforeRequest(handler func(w http.ResponseWriter, r *http.Request) (int, error)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.addBeforeHandler(handler)
}

func (this *caesar) AddAfterRequest(handler func(w http.ResponseWriter, r *http.Request)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.addAfterHandler(handler)
}

func (this *caesar) SetErrorHandler(handler func(w http.ResponseWriter, r *http.Request, code int, err error)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.setErrorHandler(handler)
}

func (this *caesar) SetNotFoundHandler(handler func(w http.ResponseWriter, r *http.Request)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.setNotFoundHandler(handler)
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
		path, err := makeRequestURI("/", h.Path)
		if err != nil {
			return err
		}
		logger.Debugf("app handler: %s %s", h.Methods, path)
		r := this.router.HandleFunc(path, handler)
		if len(h.Methods) > 0 {
			r.Methods(h.Methods...)
		}
	}
	this.router.HandleFunc(anyPath, this.stack.notFoundHandler)
	return nil
}

// run
func (this *caesar) Run(addr string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

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

	logger.Info("Server running on ", addr)
	return http.ListenAndServe(addr, this.router)
}

func (this *caesar) SetDebug(debug bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.debug = debug
	if this.debug {
		log.SetOutputLevel(log.Ldebug)
	} else {
		log.SetOutputLevel(log.Linfo)
	}
	return
}
