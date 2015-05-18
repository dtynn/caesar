package jsonutils

import (
	"encoding/json"
	"net/http"

	"github.com/dtynn/caesar/request"
)

func GetJsonArgsFromRequest(r *http.Request, obj interface{}) error {
	c := request.GetC(r)
	return GetJsonArgsFromContext(c, obj)
}

func GetJsonArgsFromContext(c *request.C, obj interface{}) error {
	if err := json.Unmarshal(c.Body(), obj); err != nil {
		return ErrInvalidRequestBody
	}
	return nil
}
