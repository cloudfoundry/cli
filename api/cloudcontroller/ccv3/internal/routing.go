package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Params map path keys to values.  For example, if your route has the path
// pattern:
//   /person/:person_id/pets/:pet_type
// Then a correct Params map would lool like:
//   router.Params{
//     "person_id": "123",
//     "pet_type": "cats",
//   }
type Params map[string]string

// Route defines the property of a Cloud Controller V3 endpoint.
//
// Method can be one of the following:
//  GET HEAD POST PUT PATCH DELETE CONNECT OPTIONS TRACE
//
// Path conforms to Pat-style pattern matching. The following docs are taken
// from http://godoc.org/github.com/bmizerany/pat#PatternServeMux
//
// Path Patterns may contain literals or captures. Capture names start with a
// colon and consist of letters A-Z, a-z, _, and 0-9. The rest of the pattern
// matches literally. The portion of the URL matching each name ends with an
// occurrence of the character in the pattern immediately following the name,
// or a /, whichever comes first. It is possible for a name to match the empty
// string.
//
// Example pattern with one capture:
//   /hello/:name
// Will match:
//   /hello/blake
//   /hello/keith
// Will not match:
//   /hello/blake/
//   /hello/blake/foo
//   /foo
//   /foo/bar
//
// Example 2:
//    /hello/:name/
// Will match:
//   /hello/blake/
//   /hello/keith/foo
//   /hello/blake
//   /hello/keith
// Will not match:
//   /foo
//   /foo/bar
type Route struct {
	// Name is a key specifying which HTTP route the router should associate with
	// the endpoint at runtime.
	Name string
	// Method is any valid HTTP method
	Method string
	// Path contains a path pattern
	Path string
	// Resource is a key specifying which resource root the router should
	// associate with the endpoint at runtime.
	Resource string
}

// CreatePath combines the route's path pattern with a Params map
// to produce a valid path.
func (r Route) CreatePath(params Params) (string, error) {
	components := strings.Split(r.Path, "/")
	for i, c := range components {
		if len(c) == 0 {
			continue
		}
		if c[0] == ':' {
			val, ok := params[c[1:]]
			if !ok {
				return "", fmt.Errorf("missing param %s", c)
			}
			components[i] = val
		}
	}

	u, err := url.Parse(strings.Join(components, "/"))
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Router combines route and resource information in order to generate HTTP
// requests.
type Router struct {
	routes    map[string]Route
	resources map[string]string
}

// NewRouter returns a pointer to a new Router.
func NewRouter(routes []Route, resources map[string]string) *Router {
	mappedRoutes := map[string]Route{}
	for _, route := range routes {
		mappedRoutes[route.Name] = route
	}
	return &Router{
		routes:    mappedRoutes,
		resources: resources,
	}
}

// CreateRequest returns a request key'd off of the name given. The params are
// merged into the URL and body is set as the request body.
func (router Router) CreateRequest(name string, params Params, body io.Reader) (*http.Request, error) {
	route, ok := router.routes[name]
	if !ok {
		return &http.Request{}, fmt.Errorf("No route exists with the name %s", name)
	}

	uri, err := route.CreatePath(params)
	if err != nil {
		return &http.Request{}, err
	}

	resource, ok := router.resources[route.Resource]
	if !ok {
		return &http.Request{}, fmt.Errorf("No resource exists with the name %s", route.Resource)
	}

	url, err := router.urlFrom(resource, uri)
	if err != nil {
		return &http.Request{}, err
	}

	return http.NewRequest(route.Method, url, body)
}

func (Router) urlFrom(resource string, uri string) (string, error) {
	u, err := url.Parse(resource)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, uri)
	return u.String(), nil
}
