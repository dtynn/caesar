package caesar

import (
	"net/http"
)

type Blueprint struct {
	mux *http.ServeMux
}

func NewBlueprint() *Blueprint {
	return &Blueprint{
		mux: http.NewServeMux(),
	}
}

func (this *Blueprint) Register() {
	return
}

func (this *Blueprint) Get() {
	return
}

func (this *Blueprint) Post() {
	return
}

func (this *Blueprint) Delete() {
	return
}

func (this *Blueprint) Put() {
	return
}

func (this *Blueprint) BeforeRequest() {

}

func (this *Blueprint) AfterRequest() {

}

func (this *Blueprint) BeforeAppRequest() {

}

func (this *Blueprint) AfterAppRequest() {

}
