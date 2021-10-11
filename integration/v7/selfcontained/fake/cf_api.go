package fake

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

type CFAPI struct {
	server *ghttp.Server
}

type CFAPIConfig struct {
	Routes map[string]Response
}

type Response struct {
	Code int
	Body interface{}
}

func NewCFAPI() *CFAPI {
	server := ghttp.NewServer()
	return &CFAPI{
		server: server,
	}
}

func (a *CFAPI) SetConfiguration(config CFAPIConfig) {
	a.server.Reset()

	for request, response := range config.Routes {
		method, path := parseRequest(request)
		responseBytes, err := json.Marshal(response.Body)
		Expect(err).NotTo(HaveOccurred())

		a.server.RouteToHandler(method, path, ghttp.RespondWith(response.Code, responseBytes))
	}
}

func (a *CFAPI) Close() {
	a.server.Close()
}

func (a *CFAPI) URL() string {
	return a.server.URL()
}

func parseRequest(request string) (string, string) {
	fields := strings.Split(request, " ")
	Expect(fields).To(HaveLen(2))
	return fields[0], fields[1]
}
