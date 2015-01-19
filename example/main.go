package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/dtynn/caesar"
	"github.com/qiniu/log"
)

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("handler Index"))
}

func handlerA(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Millisecond)
	w.Write([]byte("handler A"))
}

func handlerB(c *caesar.C) {
	c.W.Write([]byte("handler B\n"))
	c.W.Write([]byte(fmt.Sprintf("ID: %s", c.Args["id"])))
}

func handlerBPut(w http.ResponseWriter, r *http.Request) {
	c := caesar.GetC(r)
	w.Write([]byte("handler B Default\n"))
	w.Write([]byte(fmt.Sprintf("ID: %s\n", c.Args["id"])))
	w.Write([]byte("body from c1: "))
	w.Write(c.Body())
	w.Write([]byte("\n"))

	w.Write([]byte("body from c2: "))
	w.Write(c.Body())
	w.Write([]byte("\n"))

	w.Write([]byte("body from r1: "))
	buf1 := new(bytes.Buffer)
	buf1.ReadFrom(r.Body)
	w.Write(buf1.Bytes())
	w.Write([]byte("\n"))

	w.Write([]byte("body from r2: "))
	buf2 := new(bytes.Buffer)
	buf2.ReadFrom(r.Body)
	w.Write(buf2.Bytes())
	w.Write([]byte("\n"))
}

func hanlderAny(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("handler Any"))
}

func handlerPanic(w http.ResponseWriter, r *http.Request) {
	m := map[string]http.ResponseWriter{}
	m["a"].WriteHeader(200)
	w.Write([]byte("handler panic"))
}

func errorHandler(w http.ResponseWriter, r *http.Request, code int, err error) {
	log.Info("code:", code, "err:", err)
	w.WriteHeader(code)
	w.Write([]byte("custome error handler:\n"))
	w.Write([]byte(err.Error()))
}

func beforeHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "DELETE" {
		return 405, fmt.Errorf("method not allowed")
	}
	return 0, nil
}

func main() {
	c := caesar.New()
	c.Any("/", handlerIndex)
	c.Get("/a", handlerA)
	c.Post("/b/{id}", handlerB)
	c.Put("/b/{id}", handlerBPut)
	c.Any("/any", hanlderAny)
	c.Get("/p", handlerPanic)

	c.AddBeforeRequest(beforeHandler)
	c.AddAfterRequest(caesar.TimerAfterHandler)

	c.SetErrorHandler(errorHandler)

	c.Run("127.0.0.1:50081")
}
