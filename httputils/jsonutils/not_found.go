package jsonutils

import (
	"fmt"
	"net/http"
)

func RouteNotFound(w http.ResponseWriter, r *http.Request) {
	oErr := NewAPIError(http.StatusNotFound, http.StatusNotFound, fmt.Sprintf("%s %s not found", r.Method, r.URL.Path))
	OutputJsonError(w, r, 0, oErr)
	return
}
