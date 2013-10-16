package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	testapi "testhelpers/api"
	"testing"
)

func TestFindAllInOrgByRole(t *testing.T) {
	managersResp := `{"resources": [ {"metadata": {"guid": "user-1-guid"}, "entity": {}} ] }`
	billingManagersResp := `{"resources": [
	  {"metadata": {"guid": "user-2-guid"}, "entity": {}},
	  {"metadata": {"guid": "user-3-guid"}, "entity": {}}
	]}`
	auditorsResp := `{"resources": [] }`

	ccOrgManagersEndpoint, ccOrgManagersEndpointStatus := testapi.CreateCheckableEndpoint(
		"GET", "/v2/organizations/my-org-guid/managers", nil,
		testapi.TestResponse{Status: http.StatusOK, Body: managersResp},
	)
	ccBillingManagersEndpoint, ccBillingManagersEndpointStatus := testapi.CreateCheckableEndpoint(
		"GET", "/v2/organizations/my-org-guid/billing_managers", nil,
		testapi.TestResponse{Status: http.StatusOK, Body: billingManagersResp},
	)
	ccOrgAuditorsEndpoint, ccOrgAuditorsEndpointStatus := testapi.CreateCheckableEndpoint(
		"GET", "/v2/organizations/my-org-guid/auditors", nil,
		testapi.TestResponse{Status: http.StatusOK, Body: auditorsResp},
	)

	ccEndpoints := func(res http.ResponseWriter, req *http.Request) {
		switch {
		case strings.Contains(req.RequestURI, "billing"):
			ccBillingManagersEndpoint(res, req)
		case strings.Contains(req.RequestURI, "managers"):
			ccOrgManagersEndpoint(res, req)
		case strings.Contains(req.RequestURI, "auditors"):
			ccOrgAuditorsEndpoint(res, req)
		}
	}

	filter := url.QueryEscape(`Id eq "user-1-guid"`)
	uaaManagersEndpoint, uaaManagersEndpointStatus := testapi.CreateCheckableEndpoint(
		"GET", fmt.Sprintf("/Users?attributes=id,userName&filter=%s", filter), nil,
		testapi.TestResponse{Status: http.StatusOK, Body: `{ "resources": [
          { "id": "user-1-guid", "userName": "Super user 1" }
        ]}`},
	)

	filter = url.QueryEscape(`Id eq "user-2-guid" or Id eq "user-3-guid"`)
	uaaBillingManagersEndpoint, uaaBillingManagersEndpointStatus := testapi.CreateCheckableEndpoint(
		"GET", fmt.Sprintf("/Users?attributes=id,userName&filter=%s", filter), nil,
		testapi.TestResponse{Status: http.StatusOK, Body: `{ "resources": [
          { "id": "user-2-guid", "userName": "Super user 2" },
          { "id": "user-3-guid", "userName": "Super user 3" }
        ]}`},
	)

	uaaEndpoints := func(res http.ResponseWriter, req *http.Request) {
		switch {
		case strings.Contains(req.URL.RawQuery, "user-1-guid"):
			uaaManagersEndpoint(res, req)
		case strings.Contains(req.URL.RawQuery, "user-2-guid"):
			uaaBillingManagersEndpoint(res, req)
		}
	}

	cc, uaa, repo := createUsersRepo(ccEndpoints, uaaEndpoints)
	defer cc.Close()
	defer uaa.Close()

	usersByRole, apiResponse := repo.FindAllInOrgByRole(cf.Organization{Guid: "my-org-guid"})
	assert.True(t, ccOrgManagersEndpointStatus.Called())
	assert.True(t, ccBillingManagersEndpointStatus.Called())
	assert.True(t, ccOrgAuditorsEndpointStatus.Called())
	assert.True(t, uaaManagersEndpointStatus.Called())
	assert.True(t, uaaBillingManagersEndpointStatus.Called())

	assert.True(t, apiResponse.IsSuccessful())

	expectedUser1 := cf.User{Guid: "user-1-guid", Username: "Super user 1"}
	expectedUser2 := cf.User{Guid: "user-2-guid", Username: "Super user 2"}
	expectedUser3 := cf.User{Guid: "user-3-guid", Username: "Super user 3"}

	assert.Equal(t, 1, len(usersByRole["ORG MANAGER"]))
	assert.Equal(t, expectedUser1, usersByRole["ORG MANAGER"][0])

	assert.Equal(t, 2, len(usersByRole["BILLING MANAGER"]))
	assert.Equal(t, expectedUser2, usersByRole["BILLING MANAGER"][0])
	assert.Equal(t, expectedUser3, usersByRole["BILLING MANAGER"][1])

	assert.Equal(t, 0, len(usersByRole["ORG AUDITOR"]))
}

func TestFindByUsername(t *testing.T) {
	usersResponse := `{
    "resources": [
        { "id": "my-guid", "userName": "my-full-username" }
    ]
}`

	endpoint, endpointStatus := testapi.CreateCheckableEndpoint(
		"GET",
		"/Users?attributes=id,userName&filter=userName+Eq+%22damien%2Buser1%40pivotallabs.com%22",
		nil,
		testapi.TestResponse{Status: http.StatusOK, Body: usersResponse},
	)

	_, uaa, repo := createUsersRepo(nil, endpoint)
	defer uaa.Close()

	user, apiResponse := repo.FindByUsername("damien+user1@pivotallabs.com")
	assert.True(t, endpointStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, user, cf.User{Username: "my-full-username", Guid: "my-guid"})
}

func TestFindByUsernameWhenNotFound(t *testing.T) {
	endpoint, endpointStatus := testapi.CreateCheckableEndpoint(
		"GET",
		"/Users?attributes=id,userName&filter=userName+Eq+%22my-user%22",
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

func TestSetOrgRoleToOrgManager(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, org cf.Organization) net.ApiResponse {
		return repo.SetOrgRole(user, org, "OrgManager")
	}

	testSetOrUnsetOrgRoleWithValidRole(
		t, setOrUnset, "PUT", "/v2/organizations/my-org-guid/managers/my-user-guid",
	)
}

func TestSetOrgRoleToBillingManager(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, org cf.Organization) net.ApiResponse {
		return repo.SetOrgRole(user, org, "BillingManager")
	}

	testSetOrUnsetOrgRoleWithValidRole(
		t, setOrUnset, "PUT", "/v2/organizations/my-org-guid/billing_managers/my-user-guid",
	)
}

func TestSetOrgRoleToOrgAuditor(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, org cf.Organization) net.ApiResponse {
		return repo.SetOrgRole(user, org, "OrgAuditor")
	}

	testSetOrUnsetOrgRoleWithValidRole(
		t, setOrUnset, "PUT", "/v2/organizations/my-org-guid/auditors/my-user-guid",
	)
}

func TestSetOrgRoleWithInvalidRole(t *testing.T) {
	_, _, repo := createUsersRepo(nil, nil)
	apiResponse := repo.SetOrgRole(cf.User{}, cf.Organization{}, "foo")

	assert.False(t, apiResponse.IsSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid Role")
}

func TestUnsetOrgRoleFromOrgManager(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, org cf.Organization) net.ApiResponse {
		return repo.UnsetOrgRole(user, org, "OrgManager")
	}

	testSetOrUnsetOrgRoleWithValidRole(
		t, setOrUnset, "DELETE", "/v2/organizations/my-org-guid/managers/my-user-guid",
	)
}

func TestUnsetOrgRoleFromBillingManager(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, org cf.Organization) net.ApiResponse {
		return repo.UnsetOrgRole(user, org, "BillingManager")
	}

	testSetOrUnsetOrgRoleWithValidRole(
		t, setOrUnset, "DELETE", "/v2/organizations/my-org-guid/billing_managers/my-user-guid",
	)
}

func TestUnsetOrgRoleFromOrgAuditor(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, org cf.Organization) net.ApiResponse {
		return repo.UnsetOrgRole(user, org, "OrgAuditor")
	}

	testSetOrUnsetOrgRoleWithValidRole(
		t, setOrUnset, "DELETE", "/v2/organizations/my-org-guid/auditors/my-user-guid",
	)
}

func TestUnsetOrgRoleWithInvalidRole(t *testing.T) {
	_, _, repo := createUsersRepo(nil, nil)
	apiResponse := repo.UnsetOrgRole(cf.User{}, cf.Organization{}, "foo")

	assert.False(t, apiResponse.IsSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid Role")
}

func testSetOrUnsetOrgRoleWithValidRole(
	t *testing.T,
	setOrUnset func(UserRepository, cf.User, cf.Organization) net.ApiResponse,
	verb string,
	path string) {

	ccEndpoint, ccEndpointStatus := testapi.CreateCheckableEndpoint(
		verb, path, nil, testapi.TestResponse{Status: http.StatusOK},
	)

	cc, _, repo := createUsersRepo(ccEndpoint, nil)
	defer cc.Close()

	user := cf.User{Guid: "my-user-guid"}
	org := cf.Organization{Guid: "my-org-guid"}
	apiResponse := setOrUnset(repo, user, org)

	assert.True(t, ccEndpointStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func testSetOrUnsetSpaceRoleWithValidRole(
	t *testing.T,
	setOrUnset func(UserRepository, cf.User, cf.Space) net.ApiResponse,
	verb string,
	path string) {

	ccEndpoint, ccEndpointStatus := testapi.CreateCheckableEndpoint(
		verb, path, nil, testapi.TestResponse{Status: http.StatusOK},
	)

	cc, _, repo := createUsersRepo(ccEndpoint, nil)
	defer cc.Close()

	user := cf.User{Guid: "my-user-guid"}
	space := cf.Space{Guid: "my-space-guid"}
	apiResponse := setOrUnset(repo, user, space)

	assert.True(t, ccEndpointStatus.Called())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestSetSpaceRoleToSpaceManager(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, space cf.Space) net.ApiResponse {
		return repo.SetSpaceRole(user, space, "SpaceManager")
	}

	testSetOrUnsetSpaceRoleWithValidRole(t, setOrUnset, "PUT", "/v2/spaces/my-space-guid/managers/my-user-guid")
}

func TestSetSpaceRoleToSpaceDeveloper(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, space cf.Space) net.ApiResponse {
		return repo.SetSpaceRole(user, space, "SpaceDeveloper")
	}

	testSetOrUnsetSpaceRoleWithValidRole(t, setOrUnset, "PUT", "/v2/spaces/my-space-guid/developers/my-user-guid")
}

func TestSetSpaceRoleToSpaceAuditor(t *testing.T) {
	setOrUnset := func(repo UserRepository, user cf.User, space cf.Space) net.ApiResponse {
		return repo.SetSpaceRole(user, space, "SpaceAuditor")
	}

	testSetOrUnsetSpaceRoleWithValidRole(t, setOrUnset, "PUT", "/v2/spaces/my-space-guid/auditors/my-user-guid")
}

func TestSetSpaceRoleWithInvalidRole(t *testing.T) {
	_, _, repo := createUsersRepo(nil, nil)
	apiResponse := repo.SetSpaceRole(cf.User{}, cf.Space{}, "foo")

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
