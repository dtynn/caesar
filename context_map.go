package caesar

import (
	"net/http"
	"sync"
)

var defaultContextMap = &contextMap{
	cMap: map[*http.Request]*C{},
}

func GetC(r *http.Request) *C {
	return defaultContextMap.get(r)
}

type contextMap struct {
	mutex sync.RWMutex

	cMap map[*http.Request]*C
}

func (this *contextMap) add(c *C) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.cMap[c.Req] = c

	return
}

func (this *contextMap) del(c *C) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if _, ok := this.cMap[c.Req]; ok {
		delete(this.cMap, c.Req)
	}

	return
}

func (this *contextMap) get(r *http.Request) *C {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.cMap[r]
}
