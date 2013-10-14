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

func TestFindByUsername(t *testing.T) {
	usersResponse := `{
    "resources": [
        { "id": "my-guid", "userName": "my-full-username" }
    ]
}`

	endpoint, endpointStatus := testapi.CreateCheckableEndpoint(
		"GET",
		"/Users?attributes=id,userName&filter=userName%20Eq%20\"my-user\"",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: usersResponse},
	)

	_, uaa, repo := createUsersRepo(nil, endpoint)
	defer uaa.Close()

	user, apiResponse := repo.FindByUsername("my-user")
	assert.True(t, endpointStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, user, cf.User{Username: "my-full-username", Guid: "my-guid"})
}

func TestFindByUsernameWhenNotFound(t *testing.T) {
	endpoint, endpointStatus := testapi.CreateCheckableEndpoint(
		"GET",
		"/Users?attributes=id,userName&filter=userName%20Eq%20\"my-user\"",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	)

	_, uaa, repo := createUsersRepo(nil, endpoint)
	defer uaa.Close()

	_, apiResponse := repo.FindByUsername("my-user")
	assert.True(t, endpointStatus.Called())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

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

func TestDeleteUser(t *testing.T) {
	ccEndpoint, ccEndpointStatus := testapi.CreateCheckableEndpoint(
		"DELETE",
		"/v2/users/my-user-guid",
		nil,
		testapi.TestResponse{Status: http.StatusOK},
	)

	uaaEndpoint, uaaEndpointStatus := testapi.CreateCheckableEndpoint(
		"DELETE",
		"/Users/my-user-guid",
		nil,
		testapi.TestResponse{Status: http.StatusOK},
	)

	cc, uaa, repo := createUsersRepo(ccEndpoint, uaaEndpoint)
	defer cc.Close()
	defer uaa.Close()

	apiResponse := repo.Delete(cf.User{Guid: "my-user-guid"})
	assert.True(t, ccEndpointStatus.Called())
	assert.True(t, uaaEndpointStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestDeleteUserWhenNotFoundOnTheCloudController(t *testing.T) {
	ccEndpoint, ccEndpointStatus := testapi.CreateCheckableEndpoint(
		"DELETE",
		"/v2/users/my-user-guid",
		nil,
		testapi.TestResponse{Status: http.StatusNotFound, Body: `{
		  "code": 20003, "description": "The user could not be found"
		}`},
	)

	uaaEndpoint, uaaEndpointStatus := testapi.CreateCheckableEndpoint(
		"DELETE",
		"/Users/my-user-guid",
		nil,
		testapi.TestResponse{Status: http.StatusOK},
	)

	cc, uaa, repo := createUsersRepo(ccEndpoint, uaaEndpoint)
	defer cc.Close()
	defer uaa.Close()

	apiResponse := repo.Delete(cf.User{Guid: "my-user-guid"})
	assert.True(t, ccEndpointStatus.Called())
	assert.True(t, uaaEndpointStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

//"managers_url": "/v2/organizations/27162c62-fa9f-4999-af63-a12db5ed415b/managers",
//"billing_managers_url": "/v2/organizations/27162c62-fa9f-4999-af63-a12db5ed415b/billing_managers",
//"auditors_url": "/v2/organizations/27162c62-fa9f-4999-af63-a12db5ed415b/auditors",

//OrgManager
//BillingManager
//OrgAuditor

func TestSetOrgRoleToOrgManager(t *testing.T) {
	testSetOrgRoleWithValidRole(t, "OrgManager", "/v2/organizations/my-org-guid/managers/my-user-guid")
}

func TestSetOrgRoleToBillingManager(t *testing.T) {
	testSetOrgRoleWithValidRole(t, "BillingManager", "/v2/organizations/my-org-guid/billing_managers/my-user-guid")
}

func TestSetOrgRoleToOrgAuditor(t *testing.T) {
	testSetOrgRoleWithValidRole(t, "OrgAuditor", "/v2/organizations/my-org-guid/auditors/my-user-guid")
}

func testSetOrgRoleWithValidRole(t *testing.T, role string, path string) {
	ccEndpoint, ccEndpointStatus := testapi.CreateCheckableEndpoint(
		"PUT",
		path,
		nil,
		testapi.TestResponse{Status: http.StatusOK},
	)

	cc, _, repo := createUsersRepo(ccEndpoint, nil)
	defer cc.Close()

	apiResponse := repo.SetOrgRole(
		cf.User{Guid: "my-user-guid"},
		cf.Organization{Guid: "my-org-guid"},
		role,
	)

	assert.True(t, ccEndpointStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestSetOrgRoleWithInvalidRole(t *testing.T) {
	_, _, repo := createUsersRepo(nil, nil)
	apiResponse := repo.SetOrgRole(
		cf.User{},
		cf.Organization{},
		"foo",
	)

	assert.False(t, apiResponse.IsSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid Role")
}

func createUsersRepo(ccEndpoint http.HandlerFunc, uaaEndpoint http.HandlerFunc) (cc *httptest.Server, uaa *httptest.Server, repo UserRepository) {
	ccTarget := ""
	uaaTarget := ""

	if ccEndpoint != nil {
		cc = httptest.NewTLSServer(ccEndpoint)
		ccTarget = cc.URL
	}
	if uaaEndpoint != nil {
		uaa = httptest.NewTLSServer(uaaEndpoint)
		uaaTarget = uaa.URL
	}

	config := &configuration.Configuration{
		AccessToken:  "BEARER my_access_token",
		Target:       ccTarget,
		Organization: cf.Organization{Guid: "some-org-guid"},
	}
	ccGateway := net.NewCloudControllerGateway()
	uaaGateway := net.NewUAAGateway()
	endpointRepo := &testapi.FakeEndpointRepo{GetEndpointEndpoints: map[cf.EndpointType]string{
		cf.UaaEndpointKey: uaaTarget,
	}}
	repo = NewCloudControllerUserRepository(config, uaaGateway, ccGateway, endpointRepo)
	return
}
