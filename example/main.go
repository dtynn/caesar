package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dtynn/caesar"
	"github.com/dtynn/caesar/request"
	"github.com/qiniu/log"
)

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("handler Index"))
}

func handlerDefault(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Millisecond)
	w.Write([]byte("handler golang http type"))
}

func handlerCaesar(c *request.C) {
	c.W.Write([]byte("handler caesar type\n"))
}

func handlerSleep(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Millisecond)
	w.Write([]byte("handler sleep"))
}

func handlerRest(w http.ResponseWriter, r *http.Request) {
	c := request.GetC(r)
	w.Write([]byte("handler rest\n"))
	w.Write([]byte(fmt.Sprintf("ID: %s\n", c.Args["id"])))
}

func hanlderAny(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("handler Any"))
}

func handlerPanic(w http.ResponseWriter, r *http.Request) {
	m := map[string]http.ResponseWriter{}
	m["a"].WriteHeader(200)
	w.Write([]byte("handler panic"))
}

func errorHandlerApp(w http.ResponseWriter, r *http.Request, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte("custome error handler for app:\n"))
	w.Write([]byte(err.Error()))
}

func errorHandlerBp(w http.ResponseWriter, r *http.Request, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte("custome error handler for blueprint:\n"))
	w.Write([]byte(err.Error()))
}

func before1(w http.ResponseWriter, r *http.Request) (int, error) {
	log.Info("before app")
	return 0, nil
}

func before2(w http.ResponseWriter, r *http.Request) (int, error) {
	log.Info("before bp1")
	return 0, fmt.Errorf("err in before")
}

func before3(w http.ResponseWriter, r *http.Request) (int, error) {
	log.Info("before bp2")
	return 0, nil
}

func after1(w http.ResponseWriter, r *http.Request) {
	log.Info("after app")
	return
}

func after2(w http.ResponseWriter, r *http.Request) {
	log.Info("after bp1")
	return
}

func after3(w http.ResponseWriter, r *http.Request) {
	log.Info("after bp2")
	return
}

func main() {
	log.SetOutputLevel(log.Ldebug)
	bp1, _ := caesar.NewBlueprint("/bp1/")
	bp1.Any("/", handlerIndex)
	bp1.Get("/d", handlerDefault)
	bp1.Get("/c", handlerCaesar)
	bp1.Get("/s", handlerSleep)
	bp1.Get("/r/{id}", handlerRest)
	bp1.Get("/p", handlerPanic)
	bp1.AddBeforeRequest(before2)
	bp1.AddAfterRequest(after2)

	bp2, _ := caesar.NewBlueprint("/bp2")
	bp2.Any("", handlerIndex)
	bp2.Get("d", handlerDefault)
	bp2.Get("c", handlerCaesar)
	bp2.Get("s", handlerSleep)
	bp2.Get("r/{id}", handlerRest)
	bp2.Get("p", handlerPanic)
	bp2.AddBeforeRequest(before3)
	bp2.AddAfterRequest(after3)
	bp2.SetErrorHandler(errorHandlerBp)

	c := caesar.New()
	c.Any("/", handlerIndex)
	c.Get("/d", handlerDefault)
	c.Get("/c", handlerCaesar)
	c.Get("/s", handlerSleep)
	c.Post("/r/{id}", handlerRest)
	c.Get("/p", handlerPanic)
	c.Any("/any", hanlderAny)

	c.AddBeforeRequest(before1)
	c.AddAfterRequest(request.TimerAfterHandler)
	c.AddAfterRequest(after1)

	c.SetErrorHandler(errorHandlerApp)

	c.RegisterBlueprint(bp1)
	c.RegisterBlueprint(bp2)

	c.Run("127.0.0.1:50081")
}
