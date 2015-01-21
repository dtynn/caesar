package request

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type abort struct {
	code int
	err  error
}

type C struct {
	Req *http.Request
	W   http.ResponseWriter

	Logger *xlogger

	Args map[string]string
	g    map[string]interface{}

	abort *abort
	start time.Time

	mutex sync.RWMutex
}

func NewContext(w http.ResponseWriter, r *http.Request) *C {
	body := newReqBody(r.Body)
	r.Body = body
	c := &C{
		Req: r,
		W:   w,

		Logger: newXLogger(w, r),

		Args: mux.Vars(r),
		g:    map[string]interface{}{},

		start: time.Now(),
	}
	defaultContextMap.add(c)
	return c
}

func (this *C) Set(key string, val interface{}) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.g[key] = val
	return
}

func (this *C) Get(key string) interface{} {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.g[key]
}

func (this *C) Body() []byte {
	return this.Req.Body.(*reqBody).Bytes()
}

func (this *C) Abort(code int, err error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.abort != nil {
		return
	}

	if code == 0 && err == nil {
		return
	}

	this.abort = &abort{
		code: code,
		err:  err,
	}
	return
}

func (this *C) Error() (int, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	if this.abort == nil {
		return 0, nil
	}

	return this.abort.code, this.abort.err
}
