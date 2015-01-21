package caesar

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"

	"github.com/dtynn/caesar/request"
)

var typeDefaultHandler = reflect.TypeOf(func(w http.ResponseWriter, r *http.Request) {})
var typeCaesarHandler = reflect.TypeOf(func(c *request.C) {})

func inMakerForDefaultHandler(w http.ResponseWriter, r *http.Request) []reflect.Value {
	return []reflect.Value{
		reflect.ValueOf(w),
		reflect.ValueOf(r),
	}
}

func inMakerForCaesarHandler(w http.ResponseWriter, r *http.Request) []reflect.Value {
	return []reflect.Value{
		reflect.ValueOf(request.GetC(r)),
	}
}

func handlerMaker(f interface{},
	beforeHandlers []func(w http.ResponseWriter, r *http.Request) (int, error),
	afterHandlers []func(w http.ResponseWriter, r *http.Request),
	errHandler func(w http.ResponseWriter, r *http.Request, code int, err error)) (func(rw http.ResponseWriter, req *http.Request), error) {

	if f == nil {
		return nil, fmt.Errorf("handler func must not be nil")
	}
	typ := reflect.TypeOf(f)
	val := reflect.ValueOf(f)

	var inMaker func(w http.ResponseWriter, r *http.Request) []reflect.Value

	switch typ {
	case typeDefaultHandler:
		inMaker = inMakerForDefaultHandler
	case typeCaesarHandler:
		inMaker = inMakerForCaesarHandler
	default:
		return nil, fmt.Errorf("unexpected type of handler function")
	}

	return func(rw http.ResponseWriter, req *http.Request) {
		c := request.NewContext(rw, req)
		w := request.NewResponseWriter(rw, req)

		defer func() {
			for _, after := range afterHandlers {
				after(w, req)
			}

			if code, err := c.Error(); code != 0 || err != nil {
				w.Reset()
				errHandler(w, req, code, err)
			}

			if p := recover(); p != nil {
				c.Logger.Errorf("PANIC: %v\n%s", p, string(debug.Stack()))
				w.Reset()
				errHandler(w, req, 599, fmt.Errorf("internal error"))
			}

			request.DelC(c)
			w.Output()
		}()

		for _, before := range beforeHandlers {
			if code, err := before(w, req); code != 0 || err != nil {
				if code <= 0 {
					code = 500
				}

				errHandler(w, req, code, err)
				return
			}
		}

		in := inMaker(w, req)

		val.Call(in)
		return
	}, nil
}

func beforeHandlersMaker(appBefores, bpBefores []func(w http.ResponseWriter, r *http.Request) (int, error)) []func(w http.ResponseWriter, r *http.Request) (int, error) {
	if bpBefores == nil {
		return appBefores
	}

	size := len(appBefores) + len(bpBefores)
	befores := make([]func(w http.ResponseWriter, r *http.Request) (int, error), size)
	n := 0
	for _, h := range appBefores {
		befores[n] = h
		n++
	}
	for _, h := range bpBefores {
		befores[n] = h
		n++
	}
	return befores
}

func afterHandlersMaker(appAfters, bpAfters []func(w http.ResponseWriter, r *http.Request)) []func(w http.ResponseWriter, r *http.Request) {
	if bpAfters == nil {
		return appAfters
	}

	size := len(appAfters) + len(bpAfters)
	afters := make([]func(w http.ResponseWriter, r *http.Request), size)
	n := 0
	for _, h := range bpAfters {
		afters[n] = h
		n++
	}
	for _, h := range appAfters {
		afters[n] = h
		n++
	}
	return afters
}

func errHanlderPicker(appErrHandler, bpErrHandler func(w http.ResponseWriter, r *http.Request, code int, err error)) func(w http.ResponseWriter, r *http.Request, code int, err error) {
	if bpErrHandler != nil {
		return bpErrHandler
	}
	return appErrHandler
}

func notFoundHanlderPicker(appNotFoundHandler, bpNotFoundHandler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	if bpNotFoundHandler != nil {
		return bpNotFoundHandler
	}
	return appNotFoundHandler
}
