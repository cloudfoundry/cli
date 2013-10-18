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
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

func TestUserRepoFindAll(t *testing.T) {
	ccFirstPageReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/users",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
			"next_url": "/v2/users?page=2",
			"resources": [
				{"metadata": {"guid": "user-1-guid"}, "entity": {"admin":true}},
				{"metadata": {"guid": "user-2-guid"}, "entity": {"admin":false}}
			]
		}`},
	})
	ccSecondPageReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/users",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
			"next_url": "",
			"resources": [
				{"metadata": {"guid": "user-3-guid"}, "entity": {"admin":true}}
			]
		}`},
	})

	uaaResp := `{ "resources": [
          { "id": "user-1-guid", "userName": "Super user 1" },
          { "id": "user-2-guid", "userName": "Super user 2" },
          { "id": "user-3-guid", "userName": "Super user 3" }
        ]}`
	filter := `Id eq "user-1-guid" or Id eq "user-2-guid" or Id eq "user-3-guid"`

	uaaReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     fmt.Sprintf("/Users?attributes=id,userName&filter=%s", url.QueryEscape(filter)),
		Response: testnet.TestResponse{Status: http.StatusOK, Body: uaaResp},
	})

	cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t,
		[]testnet.TestRequest{ccFirstPageReq, ccSecondPageReq},
		[]testnet.TestRequest{uaaReq})

	defer cc.Close()
	defer uaa.Close()

	users, apiResponse := repo.FindAll()

	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, 3, len(users))

	expectedUser1 := cf.User{Guid: "user-1-guid", Username: "Super user 1", IsAdmin: true}
	expectedUser2 := cf.User{Guid: "user-2-guid", Username: "Super user 2", IsAdmin: false}
	expectedUser3 := cf.User{Guid: "user-3-guid", Username: "Super user 3", IsAdmin: true}
	assert.Equal(t, expectedUser1, users[0])
	assert.Equal(t, expectedUser2, users[1])
	assert.Equal(t, expectedUser3, users[2])
}

func createUsersByRoleEndpoints(rolePaths []string) (ccReqs []testnet.TestRequest, uaaReqs []testnet.TestRequest) {
	roleResponses := []string{
		`{"resources": [ {"metadata": {"guid": "user-1-guid"}, "entity": {}} ] }`,
		`{"resources": [
	  		{"metadata": {"guid": "user-2-guid"}, "entity": {}},
	  		{"metadata": {"guid": "user-3-guid"}, "entity": {}}
		]}`,
		`{"resources": [] }`,
	}

	for index, resp := range roleResponses {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     rolePaths[index],
			Response: testnet.TestResponse{Status: http.StatusOK, Body: resp},
		})
		ccReqs = append(ccReqs, req)
	}

	uaaRoleResponses := []string{
		`{ "resources": [ { "id": "user-1-guid", "userName": "Super user 1" }]}`,
		`{ "resources": [
          { "id": "user-2-guid", "userName": "Super user 2" },
          { "id": "user-3-guid", "userName": "Super user 3" }
        ]}`,
	}
	filters := []string{
		`Id eq "user-1-guid"`,
		`Id eq "user-2-guid" or Id eq "user-3-guid"`,
	}

	for index, resp := range uaaRoleResponses {
		path := fmt.Sprintf(
			"/Users?attributes=id,userName&filter=%s",
			url.QueryEscape(filters[index]),
		)
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "GET",
			Path:     path,
			Response: testnet.TestResponse{Status: http.StatusOK, Body: resp},
		})
		uaaReqs = append(uaaReqs, req)
	}
	return
}

func TestFindAllInOrgByRole(t *testing.T) {
	rolePaths := []string{
		"/v2/organizations/my-org-guid/managers",
		"/v2/organizations/my-org-guid/billing_managers",
		"/v2/organizations/my-org-guid/auditors",
	}
	ccReqs, uaaReqs := createUsersByRoleEndpoints(rolePaths)

	cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t, ccReqs, uaaReqs)
	defer cc.Close()
	defer uaa.Close()

	usersByRole, apiResponse := repo.FindAllInOrgByRole(cf.Organization{Guid: "my-org-guid"})

	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
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

func TestFindAllInSpaceByRole(t *testing.T) {
	rolePaths := []string{
		"/v2/spaces/my-space-guid/managers",
		"/v2/spaces/my-space-guid/developers",
		"/v2/spaces/my-space-guid/auditors",
	}
	ccReqs, uaaReqs := createUsersByRoleEndpoints(rolePaths)

	cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t, ccReqs, uaaReqs)
	defer cc.Close()
	defer uaa.Close()

	usersByRole, apiResponse := repo.FindAllInSpaceByRole(cf.Space{Guid: "my-space-guid"})

	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	expectedUser1 := cf.User{Guid: "user-1-guid", Username: "Super user 1"}
	expectedUser2 := cf.User{Guid: "user-2-guid", Username: "Super user 2"}
	expectedUser3 := cf.User{Guid: "user-3-guid", Username: "Super user 3"}

	assert.Equal(t, 1, len(usersByRole["SPACE MANAGER"]))
	assert.Equal(t, expectedUser1, usersByRole["SPACE MANAGER"][0])

	assert.Equal(t, 2, len(usersByRole["SPACE DEVELOPER"]))
	assert.Equal(t, expectedUser2, usersByRole["SPACE DEVELOPER"][0])
	assert.Equal(t, expectedUser3, usersByRole["SPACE DEVELOPER"][1])

	assert.Equal(t, 0, len(usersByRole["SPACE AUDITOR"]))
}

func TestFindByUsername(t *testing.T) {
	usersResponse := `{ "resources": [
        { "id": "my-guid", "userName": "my-full-username" }
    ]}`

	uaaReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/Users?attributes=id,userName&filter=userName+Eq+%22damien%2Buser1%40pivotallabs.com%22",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: usersResponse},
	})

	uaa, handler, repo := createUsersRepoWithoutCCEndpoints(t, []testnet.TestRequest{uaaReq})
	defer uaa.Close()

	user, apiResponse := repo.FindByUsername("damien+user1@pivotallabs.com")
	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
	assert.Equal(t, user, cf.User{Username: "my-full-username", Guid: "my-guid"})
}

func TestFindByUsernameWhenNotFound(t *testing.T) {
	uaaReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "GET",
		Path:     "/Users?attributes=id,userName&filter=userName+Eq+%22my-user%22",
		Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
	})

	uaa, handler, repo := createUsersRepoWithoutCCEndpoints(t, []testnet.TestRequest{uaaReq})
	defer uaa.Close()

	_, apiResponse := repo.FindByUsername("my-user")
	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, apiResponse.IsError())
	assert.True(t, apiResponse.IsNotFound())
}

func TestCreateUser(t *testing.T) {
	ccReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "POST",
		Path:     "/v2/users",
		Matcher:  testnet.RequestBodyMatcher(`{"guid":"my-user-guid"}`),
		Response: testnet.TestResponse{Status: http.StatusCreated},
	})

	uaaReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "POST",
		Path:   "/Users",
		Matcher: testnet.RequestBodyMatcher(`{
				"userName":"my-user",
				"emails":[{"value":"my-user"}],
				"password":"my-password",
				"name":{
					"givenName":"my-user",
					"familyName":"my-user"}
				}`),
		Response: testnet.TestResponse{
			Status: http.StatusCreated,
			Body:   `{"id":"my-user-guid"}`,
		},
	})

	cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t, []testnet.TestRequest{ccReq}, []testnet.TestRequest{uaaReq})
	defer cc.Close()
	defer uaa.Close()

	user := cf.User{
		Username: "my-user",
		Password: "my-password",
	}
	apiResponse := repo.Create(user)
	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
	assert.False(t, apiResponse.IsNotSuccessful())
}

func TestDeleteUser(t *testing.T) {
	ccReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/v2/users/my-user-guid",
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	uaaReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/Users/my-user-guid",
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t, []testnet.TestRequest{ccReq}, []testnet.TestRequest{uaaReq})
	defer cc.Close()
	defer uaa.Close()

	apiResponse := repo.Delete(cf.User{Guid: "my-user-guid"})
	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestDeleteUserWhenNotFoundOnTheCloudController(t *testing.T) {
	ccReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "DELETE",
		Path:   "/v2/users/my-user-guid",
		Response: testnet.TestResponse{Status: http.StatusNotFound, Body: `{
		  "code": 20003, "description": "The user could not be found"
		}`},
	})

	uaaReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     "/Users/my-user-guid",
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t, []testnet.TestRequest{ccReq}, []testnet.TestRequest{uaaReq})
	defer cc.Close()
	defer uaa.Close()

	apiResponse := repo.Delete(cf.User{Guid: "my-user-guid"})
	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
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
	repo := createUsersRepoWithoutEndpoints()
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
	repo := createUsersRepoWithoutEndpoints()
	apiResponse := repo.UnsetOrgRole(cf.User{}, cf.Organization{}, "foo")

	assert.False(t, apiResponse.IsSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid Role")
}

func testSetOrUnsetOrgRoleWithValidRole(t *testing.T,
	setOrUnset func(UserRepository, cf.User, cf.Organization) net.ApiResponse,
	verb string,
	path string) {

	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   verb,
		Path:     path,
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	cc, handler, repo := createUsersRepoWithoutUAAEndpoints(t, []testnet.TestRequest{req})
	defer cc.Close()

	user := cf.User{Guid: "my-user-guid"}
	org := cf.Organization{Guid: "my-org-guid"}
	apiResponse := setOrUnset(repo, user, org)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func testSetOrUnsetSpaceRoleWithValidRole(t *testing.T,
	setOrUnset func(UserRepository, cf.User, cf.Space) net.ApiResponse,
	verb string,
	path string) {

	reqs := []testnet.TestRequest{}

	if verb == "PUT" {
		addToOrgReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/organizations/my-space-org-guid/users/my-user-guid",
			Response: testnet.TestResponse{Status: http.StatusOK},
		})
		reqs = append(reqs, addToOrgReq)
	}

	setOrUnsetReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   verb,
		Path:     path,
		Response: testnet.TestResponse{Status: http.StatusOK},
	})
	reqs = append(reqs, setOrUnsetReq)

	cc, handler, repo := createUsersRepoWithoutUAAEndpoints(t, reqs)
	defer cc.Close()

	user := cf.User{Guid: "my-user-guid"}
	space := cf.Space{Guid: "my-space-guid", Organization: cf.Organization{Guid: "my-space-org-guid"}}
	apiResponse := setOrUnset(repo, user, space)

	assert.True(t, handler.AllRequestsCalled())
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
	repo := createUsersRepoWithoutEndpoints()
	apiResponse := repo.SetSpaceRole(cf.User{}, cf.Space{}, "foo")

	assert.False(t, apiResponse.IsSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid Role")
}

func createUsersRepoWithoutEndpoints() (repo UserRepository) {
	_, _, _, _, repo = createUsersRepo(nil, []testnet.TestRequest{}, []testnet.TestRequest{})
	return
}

func createUsersRepoWithoutUAAEndpoints(t *testing.T, ccReqs []testnet.TestRequest) (cc *httptest.Server, ccHandler *testnet.TestHandler, repo UserRepository) {
	cc, ccHandler, _, _, repo = createUsersRepo(t, ccReqs, []testnet.TestRequest{})
	return
}

func createUsersRepoWithoutCCEndpoints(t *testing.T, uaaReqs []testnet.TestRequest) (uaa *httptest.Server, uaaHandler *testnet.TestHandler, repo UserRepository) {
	_, _, uaa, uaaHandler, repo = createUsersRepo(t, []testnet.TestRequest{}, uaaReqs)
	return
}

func createUsersRepo(t *testing.T, ccReqs []testnet.TestRequest, uaaReqs []testnet.TestRequest) (cc *httptest.Server,
	ccHandler *testnet.TestHandler, uaa *httptest.Server, uaaHandler *testnet.TestHandler, repo UserRepository) {

	ccTarget := ""
	uaaTarget := ""

	if len(ccReqs) > 0 {
		cc, ccHandler = testnet.NewTLSServer(t, ccReqs)
		ccTarget = cc.URL
	}
	if len(uaaReqs) > 0 {
		uaa, uaaHandler = testnet.NewTLSServer(t, uaaReqs)
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
