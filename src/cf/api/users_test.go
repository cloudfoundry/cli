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

func createUsersByRoleEndpoints(rolePath string) (ccReqs []testnet.TestRequest, uaaReqs []testnet.TestRequest) {
	nextUrl := rolePath + "?page=2"

	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   rolePath,
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: fmt.Sprintf(`{
				"next_url": "%s",
				"resources": [ {"metadata": {"guid": "user-1-guid"}, "entity": {}} ]
			}`, nextUrl)},
	})

	secondReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   nextUrl,
		Response: testnet.TestResponse{
			Status: http.StatusOK,
			Body: `{
				"resources": [ {"metadata": {"guid": "user-2-guid"}, "entity": {}}, {"metadata": {"guid": "user-3-guid"}, "entity": {}} ]
			}`},
	})

	ccReqs = append(ccReqs, req, secondReq)

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

func TestListUsersInOrgForRole(t *testing.T) {
	ccReqs, uaaReqs := createUsersByRoleEndpoints("/v2/organizations/my-org-guid/managers")

	cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t, ccReqs, uaaReqs)
	defer cc.Close()
	defer uaa.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	usersChan, statusChan := repo.ListUsersInOrgForRole("my-org-guid", cf.ORG_MANAGER, stopChan)

	users := []cf.UserFields{}
	for chunk := range usersChan {
		users = append(users, chunk...)
	}
	apiResponse := <-statusChan

	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, len(users), 3)
	assert.Equal(t, users[0].Guid, "user-1-guid")
	assert.Equal(t, users[0].Username, "Super user 1")
	assert.Equal(t, users[1].Guid, "user-2-guid")
	assert.Equal(t, users[1].Username, "Super user 2")
}

func TestListUsersInSpaceForRole(t *testing.T) {
	ccReqs, uaaReqs := createUsersByRoleEndpoints("/v2/spaces/my-space-guid/managers")

	cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t, ccReqs, uaaReqs)
	defer cc.Close()
	defer uaa.Close()

	stopChan := make(chan bool)
	defer close(stopChan)
	usersChan, statusChan := repo.ListUsersInSpaceForRole("my-space-guid", cf.SPACE_MANAGER, stopChan)

	users := []cf.UserFields{}
	for chunk := range usersChan {
		users = append(users, chunk...)
	}
	apiResponse := <-statusChan

	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())

	assert.Equal(t, len(users), 3)
	assert.Equal(t, users[0].Guid, "user-1-guid")
	assert.Equal(t, users[0].Username, "Super user 1")
	assert.Equal(t, users[1].Guid, "user-2-guid")
	assert.Equal(t, users[1].Username, "Super user 2")
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

	expectedUserFields := cf.UserFields{}
	expectedUserFields.Username = "my-full-username"
	expectedUserFields.Guid = "my-guid"
	assert.Equal(t, user, expectedUserFields)
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

	apiResponse := repo.Create("my-user", "my-password")
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

	apiResponse := repo.Delete("my-user-guid")
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

	apiResponse := repo.Delete("my-user-guid")
	assert.True(t, ccHandler.AllRequestsCalled())
	assert.True(t, uaaHandler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestSetOrgRoleToOrgManager(t *testing.T) {
	testSetOrgRoleWithValidRole(t, "OrgManager", "/v2/organizations/my-org-guid/managers/my-user-guid")
}

func TestSetOrgRoleToBillingManager(t *testing.T) {
	testSetOrgRoleWithValidRole(t, "BillingManager", "/v2/organizations/my-org-guid/billing_managers/my-user-guid")
}

func TestSetOrgRoleToOrgAuditor(t *testing.T) {
	testSetOrgRoleWithValidRole(t, "OrgAuditor", "/v2/organizations/my-org-guid/auditors/my-user-guid")
}

func TestSetOrgRoleWithInvalidRole(t *testing.T) {
	repo := createUsersRepoWithoutEndpoints()
	apiResponse := repo.SetOrgRole("user-guid", "org-guid", "foo")

	assert.False(t, apiResponse.IsSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid Role")
}

func testSetOrgRoleWithValidRole(t *testing.T, role string, path string) {

	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     path,
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	userReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/organizations/my-org-guid/users/my-user-guid",
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	cc, handler, repo := createUsersRepoWithoutUAAEndpoints(t, []testnet.TestRequest{req, userReq})
	defer cc.Close()

	apiResponse := repo.SetOrgRole("my-user-guid", "my-org-guid", role)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestUnsetOrgRoleFromOrgManager(t *testing.T) {
	testUnsetOrgRoleWithValidRole(t, "OrgManager", "/v2/organizations/my-org-guid/managers/my-user-guid")
}

func TestUnsetOrgRoleFromBillingManager(t *testing.T) {
	testUnsetOrgRoleWithValidRole(t, "BillingManager", "/v2/organizations/my-org-guid/billing_managers/my-user-guid")
}

func TestUnsetOrgRoleFromOrgAuditor(t *testing.T) {
	testUnsetOrgRoleWithValidRole(t, "OrgAuditor", "/v2/organizations/my-org-guid/auditors/my-user-guid")
}

func TestUnsetOrgRoleWithInvalidRole(t *testing.T) {
	repo := createUsersRepoWithoutEndpoints()
	apiResponse := repo.UnsetOrgRole("user-guid", "org-guid", "foo")

	assert.False(t, apiResponse.IsSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid Role")
}

func testUnsetOrgRoleWithValidRole(t *testing.T, role string, path string) {
	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "DELETE",
		Path:     path,
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	cc, handler, repo := createUsersRepoWithoutUAAEndpoints(t, []testnet.TestRequest{req})
	defer cc.Close()

	apiResponse := repo.UnsetOrgRole("my-user-guid", "my-org-guid", role)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
}

func TestSetSpaceRoleToSpaceManager(t *testing.T) {
	testSetSpaceRoleWithValidRole(t, "SpaceManager", "/v2/spaces/my-space-guid/managers/my-user-guid")
}

func TestSetSpaceRoleToSpaceDeveloper(t *testing.T) {
	testSetSpaceRoleWithValidRole(t, "SpaceDeveloper", "/v2/spaces/my-space-guid/developers/my-user-guid")
}

func TestSetSpaceRoleToSpaceAuditor(t *testing.T) {
	testSetSpaceRoleWithValidRole(t, "SpaceAuditor", "/v2/spaces/my-space-guid/auditors/my-user-guid")
}

func TestSetSpaceRoleWithInvalidRole(t *testing.T) {
	repo := createUsersRepoWithoutEndpoints()
	apiResponse := repo.SetSpaceRole("user-guid", "space-guid", "org-guid", "foo")

	assert.False(t, apiResponse.IsSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid Role")
}

func testSetSpaceRoleWithValidRole(t *testing.T, role string, path string) {

	addToOrgReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     "/v2/organizations/my-org-guid/users/my-user-guid",
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	setRoleReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:   "PUT",
		Path:     path,
		Response: testnet.TestResponse{Status: http.StatusOK},
	})

	cc, handler, repo := createUsersRepoWithoutUAAEndpoints(t, []testnet.TestRequest{addToOrgReq, setRoleReq})
	defer cc.Close()

	apiResponse := repo.SetSpaceRole("my-user-guid", "my-space-guid", "my-org-guid", role)

	assert.True(t, handler.AllRequestsCalled())
	assert.True(t, apiResponse.IsSuccessful())
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
	org := cf.OrganizationFields{}
	org.Guid = "some-org-guid"
	config := &configuration.Configuration{
		AccessToken:        "BEARER my_access_token",
		Target:             ccTarget,
		OrganizationFields: org,
	}
	ccGateway := net.NewCloudControllerGateway()
	uaaGateway := net.NewUAAGateway()
	endpointRepo := &testapi.FakeEndpointRepo{GetEndpointEndpoints: map[cf.EndpointType]string{
		cf.UaaEndpointKey: uaaTarget,
	}}
	repo = NewCloudControllerUserRepository(config, uaaGateway, ccGateway, endpointRepo)
	return
}
