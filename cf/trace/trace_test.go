package trace_test

import (
	. "code.cloudfoundry.org/cli/cf/trace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("trace", func() {
	Describe("Sanitize", func() {
		It("hides the authorization token header", func() {
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

			Expect(Sanitize(request)).To(Equal(expected))
		})

		Describe("hiding passwords in the body of requests", func() {
			It("hides passwords in query args", func() {
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
				Expect(Sanitize(request)).To(Equal(expected))
			})

			It("hides passwords in the JSON-formatted request body", func() {
				request := `
REQUEST: [2014-03-07T10:53:36-08:00]
PUT /Users/user-guid-goes-here/password HTTP/1.1

{"password":"stanleysPasswordIsCool","oldPassword":"stanleypassword!"}
`

				expected := `
REQUEST: [2014-03-07T10:53:36-08:00]
PUT /Users/user-guid-goes-here/password HTTP/1.1

{"password":"[PRIVATE DATA HIDDEN]","oldPassword":"[PRIVATE DATA HIDDEN]"}
`

				Expect(Sanitize(request)).To(Equal(expected))
			})

			It("hides password containing \" in the JSON-formatted request body", func() {
				request := `
REQUEST: [2014-03-07T10:53:36-08:00]
PUT /Users/user-guid-goes-here/password HTTP/1.1

{"password":"stanleys\"PasswordIsCool","oldPassword":"stanleypassword!"}
`

				expected := `
REQUEST: [2014-03-07T10:53:36-08:00]
PUT /Users/user-guid-goes-here/password HTTP/1.1

{"password":"[PRIVATE DATA HIDDEN]","oldPassword":"[PRIVATE DATA HIDDEN]"}
`

				Expect(Sanitize(request)).To(Equal(expected))
			})

			It("hides create-user passwords", func() {
				request := `
REQUEST: [2014-03-07T12:15:08-08:00]
POST /Users HTTP/1.1
{
	"userName": "jiro",
	"emails": [{"value":"jiro"}],
	"password": "leansushi",
	"name": {"givenName":"jiro", "familyName":"jiro"}
}
`
				expected := `
REQUEST: [2014-03-07T12:15:08-08:00]
POST /Users HTTP/1.1
{
	"userName": "jiro",
	"emails": [{"value":"jiro"}],
	"password":"[PRIVATE DATA HIDDEN]",
	"name": {"givenName":"jiro", "familyName":"jiro"}
}
`
				Expect(Sanitize(request)).To(Equal(expected))
			})
		})

		It("hides oauth tokens in the body of requests", func() {
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

{"access_token":"[PRIVATE DATA HIDDEN]","token_type":"[PRIVATE DATA HIDDEN]","refresh_token":"[PRIVATE DATA HIDDEN]","expires_in":43199,"scope":"cloud_controller.read cloud_controller.write openid password.write","jti":"c6a7c136-6497-4faf-8799-4c42e1f3c6f5"}
`

			Expect(Sanitize(response)).To(Equal(expected))
		})

		It("hides service auth tokens in the request body", func() {
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

			Expect(Sanitize(response)).To(Equal(expected))
		})

		Describe("hiding credentials in application environment variables", func() {
			It("hides the value of any key matching case-insensitive substring 'token'", func() {
				response := `
HTTP/1.1 200 OK
Content-Type: application/json;charset=utf-8

{"guid":"99fefc8e-845e-47f3-a8b1-26e8a00222d9","name":"example","environment_json":{"token":"mytoken","TOKEN":"mytoken","foo_token_bar":"mytoken","FOO_TOKEN_BAR":"mytoken"},"memory":1024,"instances":1}
`

				expected := `
HTTP/1.1 200 OK
Content-Type: application/json;charset=utf-8

{"guid":"99fefc8e-845e-47f3-a8b1-26e8a00222d9","name":"example","environment_json":{"token":"[PRIVATE DATA HIDDEN]","TOKEN":"[PRIVATE DATA HIDDEN]","foo_token_bar":"[PRIVATE DATA HIDDEN]","FOO_TOKEN_BAR":"[PRIVATE DATA HIDDEN]"},"memory":1024,"instances":1}
`

				Expect(Sanitize(response)).To(Equal(expected))
			})

			It("hides the value of any key matching case-insensitive substring 'password'", func() {
				response := `
HTTP/1.1 200 OK
Content-Type: application/json;charset=utf-8

{"guid":"99fefc8e-845e-47f3-a8b1-26e8a00222d9","name":"example","environment_json":{"password":"mypass","PASSWORD":"mypass","foo_password_bar":"mypass","FOO_PASSWORD_BAR":"mypass"},"memory":1024,"instances":1}
`

				expected := `
HTTP/1.1 200 OK
Content-Type: application/json;charset=utf-8

{"guid":"99fefc8e-845e-47f3-a8b1-26e8a00222d9","name":"example","environment_json":{"password":"[PRIVATE DATA HIDDEN]","PASSWORD":"[PRIVATE DATA HIDDEN]","foo_password_bar":"[PRIVATE DATA HIDDEN]","FOO_PASSWORD_BAR":"[PRIVATE DATA HIDDEN]"},"memory":1024,"instances":1}
`

				Expect(Sanitize(response)).To(Equal(expected))
			})
		})
	})
})
