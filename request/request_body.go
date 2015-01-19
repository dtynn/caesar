package request

import (
	"bytes"
	"io"
)

type reqBody struct {
	*bytes.Buffer
}

func newReqBody(r io.Reader) *reqBody {
	body := reqBody{new(bytes.Buffer)}
	body.ReadFrom(r)
	return &body
}

func (this *reqBody) Close() error {
	return nil
}
