package net_test

import (
	. "cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSanitizingRemovesAuthorizationToken", func() {

			request := `
REQUEST:
GET /v2/organizations HTTP/1.1
Host: api.run.pivotal.io
Accept: application/json
Authorization: bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI3NDRkNWQ1My0xODkxLTQzZjktYjNiMy1mMTQxNDZkYzQ4ZmUiLCJzdWIiOiIzM2U3ZmVkNy1iMWMyLTRjMjAtOTU0My0yMTBiMjc2ODM1MDgiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiIzM2U3ZmVkNy1iMWMyLTRjMjAtOTU0My0yMTBiMjc2ODM1MDgiLCJ1c2VyX25hbWUiOiJtZ2VoYXJkK2NsaUBwaXZvdGFsbGFicy5jb20iLCJlbWFpbCI6Im1nZWhhcmQrY2xpQHBpdm90YWxsYWJzLmNvbSIsImlhdCI6MTM3ODI0NzgxNiwiZXhwIjoxMzc4MjkxMDE2LCJpc3MiOiJodHRwczovL3VhYS5ydW4ucGl2b3RhbC5pby9vYXV0aC90b2tlbiIsImF1ZCI6WyJvcGVuaWQiLCJjbG91ZF9jb250cm9sbGVyIiwicGFzc3dvcmQiXX0.LL_QLO0SztGRENmU-9KA2WouOyPkKVENGQoUtjqrGR-UIekXMClH6fmKELzHtB69z3n9x7_jYJbvv32D-dX1J7p1CMWIDLOzXUnIUDK7cU5Q2yuYszf4v5anKiJtrKWU0_Pg87cQTZ_lWXAhdsi-bhLVR_pITxehfz7DKChjC8gh-FiuDvH5qHxxPqYHUl9jPso5OQ0y0fqZpLt8Yq23DKWaFAZehLnrhFltdQ_jSLy1QAYYZVD_HpQDf9NozKXruIvXhyIuwGj99QmUs3LSyNWecy822VqOoBtPYS6CLegMuWWlO64TJNrnZuh5YsOuW8SudJONx2wwEqARysJIHw
This is the body. Please don't get rid of me even though I contain Authorization: and some other text
	`

			expected := `
REQUEST:
GET /v2/organizations HTTP/1.1
Host: api.run.pivotal.io
Accept: application/json
Authorization: [PRIVATE DATA HIDDEN]
This is the body. Please don't get rid of me even though I contain Authorization: and some other text
	`

			assert.Equal(mr.T(), Sanitize(request), expected)
		})
		It("TestSanitizeRemovesPassword", func() {

			request := `
POST /oauth/token HTTP/1.1
Host: login.run.pivotal.io
Accept: application/json
Authorization: [PRIVATE DATA HIDDEN]
Content-Type: application/x-www-form-urlencoded

grant_type=password&password=password&scope=&username=mgehard%2Bcli%40pivotallabs.com
`

			expected := `
POST /oauth/token HTTP/1.1
Host: login.run.pivotal.io
Accept: application/json
Authorization: [PRIVATE DATA HIDDEN]
Content-Type: application/x-www-form-urlencoded

grant_type=password&password=[PRIVATE DATA HIDDEN]&scope=&username=mgehard%2Bcli%40pivotallabs.com
`
			assert.Equal(mr.T(), Sanitize(request), expected)
		})
		It("TestSanitizeRemovesOauthTokensFromBody", func() {

			response := `
HTTP/1.1 200 OK
Content-Length: 2132
Cache-Control: no-cache
Cache-Control: no-store
Cache-Control: no-store
Connection: keep-alive
Content-Type: application/json;charset=UTF-8
Date: Thu, 05 Sep 2013 16:31:43 GMT
Expires: Thu, 01 Jan 1970 00:00:00 GMT
Pragma: no-cache
Pragma: no-cache
Server: Apache-Coyote/1.1

{"access_token":"eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNmE3YzEzNi02NDk3LTRmYWYtODc5OS00YzQyZTFmM2M2ZjUiLCJzdWIiOiIzM2U3ZmVkNy1iMWMyLTRjMjAtOTU0My0yMTBiMjc2ODM1MDgiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiIzM2U3ZmVkNy1iMWMyLTRjMjAtOTU0My0yMTBiMjc2ODM1MDgiLCJ1c2VyX25hbWUiOiJtZ2VoYXJkK2NsaUBwaXZvdGFsbGFicy5jb20iLCJlbWFpbCI6Im1nZWhhcmQrY2xpQHBpdm90YWxsYWJzLmNvbSIsImlhdCI6MTM3ODM5ODcwMywiZXhwIjoxMzc4NDQxOTAzLCJpc3MiOiJodHRwczovL3VhYS5ydW4ucGl2b3RhbC5pby9vYXV0aC90b2tlbiIsImF1ZCI6WyJvcGVuaWQiLCJjbG91ZF9jb250cm9sbGVyIiwicGFzc3dvcmQiXX0.VZErs4AnXgAzEirSY1A0yV0xQItXiPqaMfpO__MBwCihEpMEtMKemvlUPn3HEKyOGINk9YzhPV30ILrBb0oPt9plCD42BLEtyr_cbeo-1zap6QuhN8YjAAKQgjNYKORSvgi9x13JrXtCGByviHVEBP39Zeum2ZoehZfClWS7YP9lUfqaIBWUDLLBQtT6AZRlbzLwH-MJ5GkH1DOkIXzuWBk0OXp4VNm38kxzLQMnOJ3aJTcWv3YBxJeIgasoQLadTPaEPLxDGeC7V6SqhGJdyyZVnGTOKLt5ict-fxDoX6CxFnT_ZuMvseSocPfS2Or0HR_FICHAv2_C_6yv_4aI7w","token_type":"bearer","refresh_token":"eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjMjM2M2E3Yi04M2MwLTRiN2ItYjg0Zi1mNTM3MTA4ZGExZmEiLCJzdWIiOiIzM2U3ZmVkNy1iMWMyLTRjMjAtOTU0My0yMTBiMjc2ODM1MDgiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiaWF0IjoxMzc4Mzk4NzAzLCJleHAiOjEzODA5OTA3MDMsImNpZCI6ImNmIiwiaXNzIjoiaHR0cHM6Ly91YWEucnVuLnBpdm90YWwuaW8vb2F1dGgvdG9rZW4iLCJncmFudF90eXBlIjoicGFzc3dvcmQiLCJ1c2VyX25hbWUiOiJtZ2VoYXJkK2NsaUBwaXZvdGFsbGFicy5jb20iLCJhdWQiOlsiY2xvdWRfY29udHJvbGxlci5yZWFkIiwiY2xvdWRfY29udHJvbGxlci53cml0ZSIsIm9wZW5pZCIsInBhc3N3b3JkLndyaXRlIl19.G8K9hVy2TGvxWEHMmVT86iQ5szMjnN0pWog2ASawpDiV8A4QODn9lJQq0G08LjjElV6wKQywAxM6eU8p32byW6RU9Tu-0iz9lW96aWSppTjsb4itbPLxsdMXLSRKOow0vuuGhwaTYx9OZIMpzNbXJVwbRRyWlhty6LVrEZp3hG37HO-N7g2oJdFZwxATaE63iL5ZnikcvKrPkBTKUGZ8OIAvsAlHQiEnbB8mfaw6Bh74ciTjOl0DYbHlZoEMQazXkLnY3INgCyErRcjtNkjRQGe6fOV4v1Wx3PAZ05gaBsAOaThgifz4Rmaf--hnrhtYI5F3g17tDmht6udZv1_C6A","expires_in":43199,"scope":"cloud_controller.read cloud_controller.write openid password.write","jti":"c6a7c136-6497-4faf-8799-4c42e1f3c6f5"}
`

			expected := `
HTTP/1.1 200 OK
Content-Length: 2132
Cache-Control: no-cache
Cache-Control: no-store
Cache-Control: no-store
Connection: keep-alive
Content-Type: application/json;charset=UTF-8
Date: Thu, 05 Sep 2013 16:31:43 GMT
Expires: Thu, 01 Jan 1970 00:00:00 GMT
Pragma: no-cache
Pragma: no-cache
Server: Apache-Coyote/1.1

{"access_token":"[PRIVATE DATA HIDDEN]","token_type":"bearer","refresh_token":"[PRIVATE DATA HIDDEN]","expires_in":43199,"scope":"cloud_controller.read cloud_controller.write openid password.write","jti":"c6a7c136-6497-4faf-8799-4c42e1f3c6f5"}
`

			assert.Equal(mr.T(), Sanitize(response), expected)
		})
		It("TestSanitizeRemovesServiceAuthTokensFromBody", func() {

			response := `
HTTP/1.1 200 OK
Content-Length: 2132
Cache-Control: no-cache
Cache-Control: no-store
Cache-Control: no-store
Connection: keep-alive
Content-Type: application/json;charset=UTF-8
Date: Thu, 05 Sep 2013 16:31:43 GMT
Expires: Thu, 01 Jan 1970 00:00:00 GMT
Pragma: no-cache
Pragma: no-cache
Server: Apache-Coyote/1.1

{"label":"some label","provider":"some provider","token":"some-token-with-stuff-in-it"}
`

			expected := `
HTTP/1.1 200 OK
Content-Length: 2132
Cache-Control: no-cache
Cache-Control: no-store
Cache-Control: no-store
Connection: keep-alive
Content-Type: application/json;charset=UTF-8
Date: Thu, 05 Sep 2013 16:31:43 GMT
Expires: Thu, 01 Jan 1970 00:00:00 GMT
Pragma: no-cache
Pragma: no-cache
Server: Apache-Coyote/1.1

{"label":"some label","provider":"some provider","token":"[PRIVATE DATA HIDDEN]"}
`

			assert.Equal(mr.T(), Sanitize(response), expected)
		})
		It("TestPrepareRedirectTransfersAuthorizationHeader", func() {

			originalReq, err := http.NewRequest("GET", "/foo", nil)
			assert.NoError(mr.T(), err)
			originalReq.Header.Set("Authorization", "my-auth-token")

			redirectReq, err := http.NewRequest("GET", "/bar", nil)
			assert.NoError(mr.T(), err)

			via := []*http.Request{originalReq}

			err = PrepareRedirect(redirectReq, via)

			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), redirectReq.Header.Get("Authorization"), "my-auth-token")
		})
		It("TestPrepareRedirectFailsAfterOneRedirect", func() {

			firstReq, err := http.NewRequest("GET", "/foo", nil)
			assert.NoError(mr.T(), err)

			secondReq, err := http.NewRequest("GET", "/manchu", nil)
			assert.NoError(mr.T(), err)

			redirectReq, err := http.NewRequest("GET", "/bar", nil)
			assert.NoError(mr.T(), err)

			via := []*http.Request{firstReq, secondReq}

			err = PrepareRedirect(redirectReq, via)

			assert.Error(mr.T(), err)
		})
	})
}
