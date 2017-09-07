package helpers

import (
	"fmt"
	"net/http"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

func StartAndTargetServerWithoutV3API() *Server {
	server := NewTLSServer()
	server.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/v2/info"),
			RespondWith(http.StatusOK, `{"api_version":"2.34.0"}`),
		),
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/"),
			RespondWith(http.StatusNotFound, "{}"),
		),
	)

	Eventually(CF("api", server.URL(), "--skip-ssl-validation")).Should(Exit(0))
	return server
}

func StartAndTargetServerWithV3Version(v3Version string) *Server {
	server := NewTLSServer()

	rootResponse := fmt.Sprintf(`{
   "links": {
      "self": {
         "href": "%[1]s"
      },
      "cloud_controller_v2": {
         "href": "%[1]s/v2",
         "meta": {
            "version": "2.94.0"
         }
      },
      "cloud_controller_v3": {
         "href": "%[1]s/v3",
         "meta": {
            "version": "%[2]s"
         }
      },
      "network_policy_v0": {
         "href": "%[1]s/networking/v0/external"
      },
      "network_policy_v1": {
         "href": "%[1]s/networking/v1/external"
      },
      "uaa": {
         "href": "%[1]s"
      },
      "logging": {
         "href": "wss://unused:443"
      },
      "app_ssh": {
         "href": "unused:2222",
         "meta": {
            "host_key_fingerprint": "unused",
            "oauth_client": "ssh-proxy"
         }
      }
   }
 }`, server.URL(), v3Version)

	v2InfoResponse := fmt.Sprintf(`{
		"api_version":"2.34.0",
		"authorization_endpoint": "%[1]s"
  }`, server.URL())

	server.RouteToHandler(http.MethodGet, "/v2/info", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(v2InfoResponse))
	})
	server.RouteToHandler(http.MethodGet, "/v3", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(`{"links":{}}`))
	})
	server.RouteToHandler(http.MethodGet, "/login", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(`{"links":{}}`))
	})
	server.RouteToHandler(http.MethodGet, "/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(rootResponse))
	})

	Eventually(CF("api", server.URL(), "--skip-ssl-validation")).Should(Exit(0))
	return server
}
