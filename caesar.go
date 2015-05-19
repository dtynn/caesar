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

	errServerClosed = fmt.Errorf("server closed")
)

const (
	AnyPath = "/{any:.*}"
)

type Config struct {
	Prefix string
}

type Caesar struct {
	mutex   sync.Mutex
	running bool
	closed  bool
	debug   bool
	quit    chan error

	cfg *Config

	blueprints []*Blueprint
	router     *mux.Router
	stack      *stack
}

func New() *Caesar {
	return &Caesar{
		blueprints: []*Blueprint{},
		router:     mux.NewRouter(),
		stack:      newAppStack(),

		quit: make(chan error, 1),

		cfg: &Config{},
	}
}

func (this *Caesar) Register(path string, f interface{}, methods ...string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.addRequestHandler(path, f, methods...)
	return
}

func (this *Caesar) Get(path string, f interface{}) {
	this.Register(path, f, "GET", "HEAD")
}

func (this *Caesar) Post(path string, f interface{}) {
	this.Register(path, f, "POST")
}

func (this *Caesar) Put(path string, f interface{}) {
	this.Register(path, f, "PUT")
}

func (this *Caesar) Delete(path string, f interface{}) {
	this.Register(path, f, "DELETE")
}

func (this *Caesar) Head(path string, f interface{}) {
	this.Register(path, f, "HEAD")
}

func (this *Caesar) Options(path string, f interface{}) {
	this.Register(path, f, "OPTIONS")
}

func (this *Caesar) Any(path string, f interface{}) {
	this.Register(path, f)
}

func (this *Caesar) RegisterBlueprint(bp *Blueprint) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if bp != nil {
		this.blueprints = append(this.blueprints, bp)
	}
}

func (this *Caesar) AddBeforeRequest(handler func(w http.ResponseWriter, r *http.Request) (int, error)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.addBeforeHandler(handler)
}

func (this *Caesar) AddAfterRequest(handler func(w http.ResponseWriter, r *http.Request)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.addAfterHandler(handler)
}

func (this *Caesar) SetErrorHandler(handler func(w http.ResponseWriter, r *http.Request, code int, err error)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.setErrorHandler(handler)
}

func (this *Caesar) SetNotFoundHandler(handler func(w http.ResponseWriter, r *http.Request)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stack.setNotFoundHandler(handler)
}

func (this *Caesar) parseHandlerFunc(f interface{}) (func(rw http.ResponseWriter, req *http.Request), error) {
	return handlerMaker(f, this.stack.beforeHandlers, this.stack.afterHandlers, this.stack.errorHandler)
}

func (this *Caesar) build() error {
	prefix := "/"
	if p := this.cfg.Prefix; p != "" {
		prefix = p
	}

	for _, h := range this.stack.requestHandlers {
		handler, err := this.parseHandlerFunc(h.Fn)
		if err != nil {
			return err
		}
		path, err := makeRequestURI(prefix, h.Path)
		if err != nil {
			return err
		}
		logger.Debugf("app handler: %s %s", h.Methods, path)
		r := this.router.HandleFunc(path, handler)
		if len(h.Methods) > 0 {
			r.Methods(h.Methods...)
		}
	}

	notFoundHandler, err := this.parseHandlerFunc(this.stack.notFoundHandler)
	if err != nil {
		return err
	}

	this.router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return nil
}

// run
func (this *Caesar) Run(addr string) error {
	this.mutex.Lock()

	if this.running {
		return fmt.Errorf("the server is already running")
	}

	if this.closed {
		return fmt.Errorf("the server is closed")
	}

	this.mutex.Unlock()

	go this.run(addr)
	err := <-this.quit
	if err != nil {
		logger.Warn(err)
	}

	return err
}

func (this *Caesar) run(addr string) {
	this.running = true
	defer func() {
		this.running = false
	}()

	// build blueprint router
	for _, bp := range this.blueprints {
		if err := bp.build(this); err != nil {
			this.quit <- err
			return
		}
	}

	// build app route
	if err := this.build(); err != nil {
		this.quit <- err
		return
	}

	logger.Info("Server running on ", addr)
	err := http.ListenAndServe(addr, this.router)
	this.quit <- err
}

func (this *Caesar) Close() {
	if this.closed {
		return
	}
	this.quit <- errServerClosed
	this.closed = true
	close(this.quit)
}

func (this *Caesar) SetDebug(debug bool) {
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

func (this *Caesar) SetConfig(cfg *Config) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if cfg != nil {
		this.cfg = cfg
	}
	return
}
