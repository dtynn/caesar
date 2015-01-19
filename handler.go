package caesar

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
)

var typeDefaultHandler = reflect.TypeOf(func(w http.ResponseWriter, r *http.Request) {})
var typeCaesarHandler = reflect.TypeOf(func(c *C) {})

func inMakerForDefaultHandler(w http.ResponseWriter, r *http.Request) []reflect.Value {
	return []reflect.Value{
		reflect.ValueOf(w),
		reflect.ValueOf(r),
	}
}

func inMakerForCaesarHandler(w http.ResponseWriter, r *http.Request) []reflect.Value {
	return []reflect.Value{
		reflect.ValueOf(GetC(r)),
	}
}

func parseHandlerFunc(f interface{}, stk *stack) (func(rw http.ResponseWriter, req *http.Request), error) {
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
		c := newContext(rw, req)
		w := newResponseWriter(rw, req)

		defer func() {
			defaultContextMap.del(c)
			if p := recover(); p != nil {
				logger.Errorf("PANIC: %v\n", p)
				logger.Errorf(string(debug.Stack()))
				w.reset()
				stk.errorHandler(w, req, 599, fmt.Errorf("internal error"))
			}
			w.output()
		}()

		for _, before := range stk.beforeHandlers {
			if code, err := before(w, req); err != nil {
				if code <= 0 {
					code = 500
				}

				stk.errorHandler(w, req, code, err)
				return
			}
		}

		in := inMaker(w, req)

		val.Call(in)

		for _, after := range stk.afterHandlers {
			after(w, req)
		}
		return
	}, nil
}
