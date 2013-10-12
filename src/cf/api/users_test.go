package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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

	cc, uaa, repo := createUsersRepo(ccEndpoint, uaaEndpoint)
	defer cc.Close()
	defer uaa.Close()

	user := cf.User{
		Username: "my-user",
		Password: "my-password",
	}
	apiResponse := repo.Create(user)
	assert.True(t, ccEndpointStatus.Called())
	assert.True(t, uaaEndpointStatus.Called())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func createUsersRepo(ccEndpoint http.HandlerFunc, uaaEndpoint http.HandlerFunc) (cc *httptest.Server, uaa *httptest.Server, repo UserRepository) {
	cc = httptest.NewTLSServer(ccEndpoint)
	uaa = httptest.NewTLSServer(uaaEndpoint)

	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       cc.URL,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	ccGateway := net.NewCloudControllerGateway()
	uaaGateway := net.NewUAAGateway()
	endpointRepo := &testapi.FakeEndpointRepo{GetEndpointEndpoints: map[cf.EndpointType]string{
		cf.UaaEndpointKey: uaa.URL,
	}}
	repo = NewCloudControllerUserRepository(config, uaaGateway, ccGateway, endpointRepo)
	return
}
