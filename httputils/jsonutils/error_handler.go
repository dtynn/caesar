package jsonutils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func OutputJsonError(w http.ResponseWriter, r *http.Request, code int, err error) {
	var oErr *APIError

	if asserted, ok := err.(*APIError); ok {
		oErr = asserted
		// priority: code > err.StatusCode
		if code > 0 {
			oErr.StatusCode = code
		}
	} else {
		oErr = NewAPIError(code, 0, err.Error())
	}

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.Encode(oErr)

	if oErr.StatusCode <= 0 {
		oErr.StatusCode = http.StatusInternalServerError
	}

	w.WriteHeader(oErr.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
	return
}
