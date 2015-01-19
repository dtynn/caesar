package caesar

type requestHandler struct {
	methods []string
	path    string
	fn      interface{}
}
