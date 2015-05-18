package jsonutils

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/dtynn/caesar/request"
)

func OutputJson(data interface{}, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if data == nil {
		return
	}

	c := request.GetC(r)

	buf := new(bytes.Buffer)

	encoder := json.NewEncoder(buf)

	err := encoder.Encode(data)
	if err != nil {
		c.Logger.Warn(err)
		OutputJsonError(w, r, ErrInternalError.StatusCode, ErrInternalError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
	return
}

func OutputJsonWithC(data interface{}, c *request.C) {
	OutputJson(data, c.W, c.Req)
}
