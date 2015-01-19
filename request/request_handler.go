package request

type RequestHandler struct {
	Methods []string
	Path    string
	Fn      interface{}
}

func NewRequestHandler(path string, fn interface{}, methods ...string) *RequestHandler {
	return &RequestHandler{
		Methods: methods,
		Path:    path,
		Fn:      fn,
	}
}
