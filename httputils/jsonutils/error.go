package jsonutils

import (
	"fmt"
	"net/http"
)

type APIError struct {
	StatusCode int    `json:"-"`
	ErrorCode  int    `json:"error_code,omitempty"`
	ErrorMsg   string `json:"error_msg"`
}

func (this *APIError) Error() string {
	return fmt.Sprintf("StatusCode: %d; ErrorCode: %d; Error: %s", this.StatusCode, this.ErrorCode, this.ErrorMsg)
}

func NewAPIError(statusCode, errorCode int, msg string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		ErrorMsg:   msg,
	}
}

var (
	ErrInvalidRequestBody = NewAPIError(http.StatusBadRequest, http.StatusBadRequest, "invalid request body")
	ErrInternalError      = NewAPIError(http.StatusInternalServerError, http.StatusInternalServerError, "internal error")
)
