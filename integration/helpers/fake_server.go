package helpers

import (
	"fmt"
	"net/http"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

const (
	DefaultV2Version string = "2.131.0"
	DefaultV3Version string = "3.66.0"
)

func StartAndTargetServerWithAPIVersions(v2Version string, v3Version string) *Server {
	server := StartServerWithAPIVersions(v2Version, v3Version)
	Eventually(CF("api", server.URL(), "--skip-ssl-validation")).Should(Exit(0))

	return server
}

func StartServerWithMinimumCLIVersion(minCLIVersion string) *Server {
	return startServerWithVersions(DefaultV2Version, DefaultV3Version, &minCLIVersion)
}

func StartServerWithAPIVersions(v2Version string, v3Version string) *Server {
	return startServerWithVersions(v2Version, v3Version, nil)
}

func startServerWithVersions(v2Version string, v3Version string, minimumCLIVersion *string) *Server {
	server := NewTLSServer()

	rootResponse := fmt.Sprintf(`{
   "links": {
      "self": {
         "href": "%[1]s"
      },
      "cloud_controller_v2": {
         "href": "%[1]s/v2",
         "meta": {
            "version": "%[2]s"
         }
      },
      "cloud_controller_v3": {
         "href": "%[1]s/v3",
         "meta": {
            "version": "%[3]s"
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
 }`, server.URL(), v2Version, v3Version)

	v2InfoResponse := struct {
		APIVersion            string  `json:"api_version"`
		AuthorizationEndpoint string  `json:"authorization_endpoint"`
		MinCLIVersion         *string `json:"min_cli_version"`
	}{
		APIVersion:            v2Version,
		AuthorizationEndpoint: server.URL(),
		MinCLIVersion:         minimumCLIVersion}

	server.RouteToHandler(http.MethodGet, "/v2/info", RespondWithJSONEncoded(http.StatusOK, v2InfoResponse))
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

	return server
}
