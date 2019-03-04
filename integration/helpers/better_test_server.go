package helpers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/onsi/gomega/ghttp"
)

type resp struct {
	status int
	body   []byte
}

var responses map[string]resp
var seenRoutes map[string]bool

func AddHandler(ser *ghttp.Server, method string, pathAndQuery string, status int, body []byte) {
	u, err := url.Parse(pathAndQuery)
	if err != nil {
		panic(err)
	}
	if len(responses) == 0 {
		responses = make(map[string]resp)
		seenRoutes = make(map[string]bool)
	}

	responses[key(method, u)] = resp{status, body}

	if !seenRoutes[key(method, u)] {
		ser.RouteToHandler(method, u.Path, func(w http.ResponseWriter, r *http.Request) {
			res := responses[key(r.Method, r.URL)]
			w.WriteHeader(res.status)
			w.Write(res.body)
		})
		seenRoutes[key(method, u)] = true
	}
}

func key(method string, url *url.URL) string {
	return strings.ToLower(method + url.String())
}
