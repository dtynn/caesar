package caesar

import (
	"net/http"

	"github.com/dtynn/caesar/request"
)

type stack struct {
	requestHandlers []*request.RequestHandler
	beforeHandlers  []func(w http.ResponseWriter, r *http.Request) (int, error)
	afterHandlers   []func(w http.ResponseWriter, r *http.Request)
	errorHandler    func(w http.ResponseWriter, r *http.Request, code int, err error)
}

func newAppStack() *stack {
	return &stack{
		requestHandlers: []*request.RequestHandler{},
		beforeHandlers:  []func(w http.ResponseWriter, r *http.Request) (int, error){},
		afterHandlers:   []func(w http.ResponseWriter, r *http.Request){},
		errorHandler:    request.DefaultErrorHandler,
	}
}

func newBpStack() *stack {
	return &stack{
		requestHandlers: []*request.RequestHandler{},
		beforeHandlers:  []func(w http.ResponseWriter, r *http.Request) (int, error){},
		afterHandlers:   []func(w http.ResponseWriter, r *http.Request){},
		errorHandler:    nil,
	}
}

func (this *stack) addRequestHandler(path string, fn interface{}, methods ...string) {
	h := request.NewRequestHandler(path, fn, methods...)
	this.requestHandlers = append(this.requestHandlers, h)
}

func (this *stack) addBeforeHandler(handler func(w http.ResponseWriter, r *http.Request) (int, error)) {
	if handler != nil {
		this.beforeHandlers = append(this.beforeHandlers, handler)
	}
}

func (this *stack) addAfterHandler(handler func(w http.ResponseWriter, r *http.Request)) {
	if handler != nil {
		this.afterHandlers = append(this.afterHandlers, handler)
	}
}

func (this *stack) setErrorHandler(handler func(w http.ResponseWriter, r *http.Request, code int, err error)) {
	if handler != nil {
		this.errorHandler = handler
	}
}
