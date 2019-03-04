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

	v3Response := fmt.Sprintf(`{"links": {
			"organizations": {
				"href": "%s/v3/organizations"
			}
		}}`, server.URL())
	server.RouteToHandler(http.MethodGet, "/v3", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(v3Response))
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

func AddMfa(server *Server, password string, mfaToken string) {
	getLoginResponse := fmt.Sprintf(`{
    "app": {
        "version": "4.28.0"
    },
    "showLoginLinks": true,
    "links": {
        "uaa": "%[1]s",
        "passwd": "/forgot_password",
        "login": "%[1]s",
        "register": "/create_account"
    },
    "zone_name": "uaa",
    "entityID": "some-host-name.example.com",
    "commit_id": "8917980",
    "idpDefinitions": {},
    "prompts": {
        "username": [
            "text",
            "Email"
        ],
        "password": [
            "password",
            "Password"
        ],
        "passcode": [
            "password",
            "Temporary Authentication Code ( Get one at %[1]s/passcode )"
        ],
        "mfaCode": [
            "password",
            "MFA Code ( Register at %[1]s )"
        ]
    },
    "timestamp": "2019-02-19T18:08:02+0000"
}`, server.URL())

	server.RouteToHandler(http.MethodGet, "/login",
		RespondWith(http.StatusOK, getLoginResponse),
	)

	server.RouteToHandler(http.MethodPost, "/oauth/token", makeMFAValidator(password, mfaToken))

}

func makeMFAValidator(password string, mfaToken string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		Expect(req.ParseForm()).To(Succeed())
		rightPassword := len(req.Form["password"]) == 1 && req.Form["password"][0] == password
		rightCode := len(req.Form["mfaCode"]) == 1 && req.Form["mfaCode"][0] == mfaToken

		if rightPassword && rightCode {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(`{
    "access_token": "some-access-token",
    "token_type": "bearer",
    "id_token": "some-id-token",
    "refresh_token": "some-refresh-token",
    "expires_in": 599,
    "scope": "openid routing.router_groups.write scim.read cloud_controller.admin uaa.user routing.router_groups.read cloud_controller.read password.write cloud_controller.write network.admin doppler.firehose scim.write",
    "jti": "66e46003f28e44c8a6582f6d6e44753f"
}`))
			return
		}
		res.WriteHeader(http.StatusUnauthorized)
	}
}

func AddLoginRoutes(s *Server) {
	s.RouteToHandler("POST", "/oauth/token", RespondWith(http.StatusOK,
		`{
			"access_token": "some-token-value",
			"expires_in": 599,
			"id_token": "some-other-token",
			"jti": "some-other-string",
			"refresh_token": "some-refresh-token",
			"scope": "openid routing.router_groups.write scim.read cloud_controller.admin uaa.user routing.router_groups.read cloud_controller.read password.write cloud_controller.write network.admin doppler.firehose scim.write",
			"token_type": "some-type"
		 }
		 `,
	))
}
