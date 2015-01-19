package caesar

import (
	"net/http"
)

type stack struct {
	requestHandlers []*requestHandler
	beforeHandlers  []func(w http.ResponseWriter, r *http.Request) (int, error)
	afterHandlers   []func(w http.ResponseWriter, r *http.Request)
	errorHandler    func(w http.ResponseWriter, r *http.Request, code int, err error)
}

func newStack() *stack {
	return &stack{
		requestHandlers: []*requestHandler{},
		beforeHandlers:  []func(w http.ResponseWriter, r *http.Request) (int, error){},
		afterHandlers:   []func(w http.ResponseWriter, r *http.Request){},
		errorHandler:    defaultErrorHandler,
	}
}

func (this *stack) addRequestHandler(path string, fn interface{}, methods ...string) {
	h := &requestHandler{
		path:    path,
		fn:      fn,
		methods: methods,
	}
	this.requestHandlers = append(this.requestHandlers, h)
}

func (this *stack) addBeforeHandler(handler func(w http.ResponseWriter, r *http.Request) (int, error)) {
	if handler != nil {
		this.beforeHandlers = append(this.beforeHandlers, handler)
	}
}

func (this *stack) setErrorHandler(handler func(w http.ResponseWriter, r *http.Request, code int, err error)) {
	if handler != nil {
		this.errorHandler = handler
	}
}

func (this *stack) addAfterHandler(handler func(w http.ResponseWriter, r *http.Request)) {
	if handler != nil {
		this.afterHandlers = append(this.afterHandlers, handler)
	}
}
