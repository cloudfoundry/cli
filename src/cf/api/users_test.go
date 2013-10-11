package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	testapi "testhelpers/api"
	"testing"
)

func TestCreateUser(t *testing.T) {
	ccEndpoint, ccEndpointStatus := testapi.CreateCheckableEndpoint(
		"POST",
		"/v2/users",
		testapi.RequestBodyMatcher(`{"guid":"my-user-guid"}`),
		testapi.TestResponse{Status: http.StatusCreated},
	)

	uaaEndpoint, uaaEndpointStatus := testapi.CreateCheckableEndpoint(
		"POST",
		"/Users",
		testapi.RequestBodyMatcher(`{
				"userName":"my-user",
				"emails":[{"value":"my-user"}],
				"password":"my-password",
				"name":{
					"givenName":"my-user",
					"familyName":"my-user"}
				}`),
		testapi.TestResponse{
			Status: http.StatusCreated,
			Body:   `{"id":"my-user-guid"}`,
		},
	)

	endpoints := func(w http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.RequestURI, "/v2/users") {
			ccEndpoint(w, req)
		} else {
			uaaEndpoint(w, req)
		}
	}

	ts, repo := createUsersRepo(endpoints)
	defer ts.Close()

	user := cf.User{
		Username: "my-user",
		Password: "my-password",
	}
	apiResponse := repo.Create(user)
	assert.True(t, ccEndpointStatus.Called())
	assert.True(t, uaaEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createUsersRepo(endpoint http.HandlerFunc) (ts *httptest.Server, repo UserRepository) {
	ts = httptest.NewTLSServer(endpoint)

	config := &configuration.Configuration{
		AccessToken:           "BEARER my_access_token",
		Target:                ts.URL,
		AuthorizationEndpoint: ts.URL,
		Organization:          cf.Organization{Guid: "some-org-guid"},
	}
	ccGateway := net.NewCloudControllerGateway()
	uaaGateway := net.NewUAAGateway()
	repo = NewCloudControllerUserRepository(config, uaaGateway, ccGateway)
	return
}
