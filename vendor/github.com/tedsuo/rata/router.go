package rata

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bmizerany/pat"
)

// Supported HTTP methods.
const (
	GET     = "GET"
	HEAD    = "HEAD"
	POST    = "POST"
	PUT     = "PUT"
	PATCH   = "PATCH"
	DELETE  = "DELETE"
	CONNECT = "CONNECT"
	OPTIONS = "OPTIONS"
	TRACE   = "TRACE"
)

// Handlers map route names to http.Handler objects.  Each Handler key must
// match a route Name in the Routes collection.
type Handlers map[string]http.Handler

// NewRouter combines a set of Routes with their corresponding Handlers to
// produce a http request multiplexer (AKA a "router").  If any route does
// not have a matching handler, an error occurs.
func NewRouter(routes Routes, handlers Handlers) (http.Handler, error) {
	p := pat.New()

	for _, route := range routes {
		handler, ok := handlers[route.Name]
		if !ok {
			return nil, fmt.Errorf("missing handler %s", route.Name)
		}

		switch method := strings.ToUpper(route.Method); method {
		case GET, HEAD, POST, PUT, PATCH, DELETE, CONNECT, OPTIONS, TRACE:
			p.Add(method, route.Path, handler)
		default:
			return nil, fmt.Errorf("invalid verb: %s", route.Method)
		}
	}

	return p, nil
}
