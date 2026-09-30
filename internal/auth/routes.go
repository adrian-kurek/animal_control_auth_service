package auth

import (
	"fmt"
	"net/http"

	"github.com/adrian-kurek/animal_control_auth_service/common/request"
)

type handler interface {
	Register(w http.ResponseWriter, r *http.Request) error
}

type Route struct {
	handler handler
}

func NewRoute(handler handler) *Route {
	return &Route{
		handler: handler,
	}
}

func (ar *Route) Setup(router *http.ServeMux) {
	prefix := "/auth"
	router.Handle(fmt.Sprintf("POST %s/register", prefix), request.Make(ar.handler.Register))
}
