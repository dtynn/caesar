package request

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type C struct {
	Req *http.Request
	W   http.ResponseWriter

	Logger *xlogger

	Args map[string]string
	g    map[string]interface{}

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
