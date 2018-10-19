package byproxy

import "net/http"

type Route struct {
	Environment string
	Handler     http.Handler
}

func NewRoute() *Route {
	return &Route{}
}

func (route *Route) Patch(environment string, handler http.Handler) {
	route.Environment = environment
	route.Handler = handler
}

func (route *Route) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	if route.Handler != nil {
		route.Handler.ServeHTTP(rw, request)
	} else {
		rw.WriteHeader(http.StatusServiceUnavailable)
	}
}
