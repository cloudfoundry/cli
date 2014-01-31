package api_test

import (
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	"net/url"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
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

func testSetOrgRoleWithValidRole(t mr.TestingT, role string, path string) {

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

func testUnsetOrgRoleWithValidRole(t mr.TestingT, role string, path string) {
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

func testSetSpaceRoleWithValidRole(t mr.TestingT, role string, path string) {

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

func createUsersRepoWithoutUAAEndpoints(t mr.TestingT, ccReqs []testnet.TestRequest) (cc *httptest.Server, ccHandler *testnet.TestHandler, repo UserRepository) {
	cc, ccHandler, _, _, repo = createUsersRepo(t, ccReqs, []testnet.TestRequest{})
	return
}

func createUsersRepoWithoutCCEndpoints(t mr.TestingT, uaaReqs []testnet.TestRequest) (uaa *httptest.Server, uaaHandler *testnet.TestHandler, repo UserRepository) {
	_, _, uaa, uaaHandler, repo = createUsersRepo(t, []testnet.TestRequest{}, uaaReqs)
	return
}

func createUsersRepo(t mr.TestingT, ccReqs []testnet.TestRequest, uaaReqs []testnet.TestRequest) (cc *httptest.Server,
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
	endpointRepo := &testapi.FakeEndpointRepo{}
	endpointRepo.UAAEndpointReturns.Endpoint = uaaTarget
	repo = NewCloudControllerUserRepository(config, uaaGateway, ccGateway, endpointRepo)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListUsersInOrgForRole", func() {
			ccReqs, uaaReqs := createUsersByRoleEndpoints("/v2/organizations/my-org-guid/managers")

			cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(mr.T(), ccReqs, uaaReqs)
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

			assert.True(mr.T(), ccHandler.AllRequestsCalled())
			assert.True(mr.T(), uaaHandler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.Equal(mr.T(), len(users), 3)
			assert.Equal(mr.T(), users[0].Guid, "user-1-guid")
			assert.Equal(mr.T(), users[0].Username, "Super user 1")
			assert.Equal(mr.T(), users[1].Guid, "user-2-guid")
			assert.Equal(mr.T(), users[1].Username, "Super user 2")
		})
		It("TestListUsersInSpaceForRole", func() {

			ccReqs, uaaReqs := createUsersByRoleEndpoints("/v2/spaces/my-space-guid/managers")

			cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(mr.T(), ccReqs, uaaReqs)
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

			assert.True(mr.T(), ccHandler.AllRequestsCalled())
			assert.True(mr.T(), uaaHandler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.Equal(mr.T(), len(users), 3)
			assert.Equal(mr.T(), users[0].Guid, "user-1-guid")
			assert.Equal(mr.T(), users[0].Username, "Super user 1")
			assert.Equal(mr.T(), users[1].Guid, "user-2-guid")
			assert.Equal(mr.T(), users[1].Username, "Super user 2")
		})
		It("TestFindByUsername", func() {

			usersResponse := `{ "resources": [
        { "id": "my-guid", "userName": "my-full-username" }
    ]}`

			uaaReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/Users?attributes=id,userName&filter=userName+Eq+%22damien%2Buser1%40pivotallabs.com%22",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: usersResponse},
			})

			uaa, handler, repo := createUsersRepoWithoutCCEndpoints(mr.T(), []testnet.TestRequest{uaaReq})
			defer uaa.Close()

			user, apiResponse := repo.FindByUsername("damien+user1@pivotallabs.com")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			expectedUserFields := cf.UserFields{}
			expectedUserFields.Username = "my-full-username"
			expectedUserFields.Guid = "my-guid"
			assert.Equal(mr.T(), user, expectedUserFields)
		})
		It("TestFindByUsernameWhenNotFound", func() {

			uaaReq := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/Users?attributes=id,userName&filter=userName+Eq+%22my-user%22",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
			})

			uaa, handler, repo := createUsersRepoWithoutCCEndpoints(mr.T(), []testnet.TestRequest{uaaReq})
			defer uaa.Close()

			_, apiResponse := repo.FindByUsername("my-user")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestCreateUser", func() {

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

			cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(mr.T(), []testnet.TestRequest{ccReq}, []testnet.TestRequest{uaaReq})
			defer cc.Close()
			defer uaa.Close()

			apiResponse := repo.Create("my-user", "my-password")
			assert.True(mr.T(), ccHandler.AllRequestsCalled())
			assert.True(mr.T(), uaaHandler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestDeleteUser", func() {

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

			cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(mr.T(), []testnet.TestRequest{ccReq}, []testnet.TestRequest{uaaReq})
			defer cc.Close()
			defer uaa.Close()

			apiResponse := repo.Delete("my-user-guid")
			assert.True(mr.T(), ccHandler.AllRequestsCalled())
			assert.True(mr.T(), uaaHandler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestDeleteUserWhenNotFoundOnTheCloudController", func() {

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

			cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(mr.T(), []testnet.TestRequest{ccReq}, []testnet.TestRequest{uaaReq})
			defer cc.Close()
			defer uaa.Close()

			apiResponse := repo.Delete("my-user-guid")
			assert.True(mr.T(), ccHandler.AllRequestsCalled())
			assert.True(mr.T(), uaaHandler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestSetOrgRoleToOrgManager", func() {

			testSetOrgRoleWithValidRole(mr.T(), "OrgManager", "/v2/organizations/my-org-guid/managers/my-user-guid")
		})
		It("TestSetOrgRoleToBillingManager", func() {

			testSetOrgRoleWithValidRole(mr.T(), "BillingManager", "/v2/organizations/my-org-guid/billing_managers/my-user-guid")
		})
		It("TestSetOrgRoleToOrgAuditor", func() {

			testSetOrgRoleWithValidRole(mr.T(), "OrgAuditor", "/v2/organizations/my-org-guid/auditors/my-user-guid")
		})
		It("TestSetOrgRoleWithInvalidRole", func() {

			repo := createUsersRepoWithoutEndpoints()
			apiResponse := repo.SetOrgRole("user-guid", "org-guid", "foo")

			assert.False(mr.T(), apiResponse.IsSuccessful())
			assert.Contains(mr.T(), apiResponse.Message, "Invalid Role")
		})
		It("TestUnsetOrgRoleFromOrgManager", func() {

			testUnsetOrgRoleWithValidRole(mr.T(), "OrgManager", "/v2/organizations/my-org-guid/managers/my-user-guid")
		})
		It("TestUnsetOrgRoleFromBillingManager", func() {

			testUnsetOrgRoleWithValidRole(mr.T(), "BillingManager", "/v2/organizations/my-org-guid/billing_managers/my-user-guid")
		})
		It("TestUnsetOrgRoleFromOrgAuditor", func() {

			testUnsetOrgRoleWithValidRole(mr.T(), "OrgAuditor", "/v2/organizations/my-org-guid/auditors/my-user-guid")
		})
		It("TestUnsetOrgRoleWithInvalidRole", func() {

			repo := createUsersRepoWithoutEndpoints()
			apiResponse := repo.UnsetOrgRole("user-guid", "org-guid", "foo")

			assert.False(mr.T(), apiResponse.IsSuccessful())
			assert.Contains(mr.T(), apiResponse.Message, "Invalid Role")
		})
		It("TestSetSpaceRoleToSpaceManager", func() {

			testSetSpaceRoleWithValidRole(mr.T(), "SpaceManager", "/v2/spaces/my-space-guid/managers/my-user-guid")
		})
		It("TestSetSpaceRoleToSpaceDeveloper", func() {

			testSetSpaceRoleWithValidRole(mr.T(), "SpaceDeveloper", "/v2/spaces/my-space-guid/developers/my-user-guid")
		})
		It("TestSetSpaceRoleToSpaceAuditor", func() {

			testSetSpaceRoleWithValidRole(mr.T(), "SpaceAuditor", "/v2/spaces/my-space-guid/auditors/my-user-guid")
		})
		It("TestSetSpaceRoleWithInvalidRole", func() {

			repo := createUsersRepoWithoutEndpoints()
			apiResponse := repo.SetSpaceRole("user-guid", "space-guid", "org-guid", "foo")

			assert.False(mr.T(), apiResponse.IsSuccessful())
			assert.Contains(mr.T(), apiResponse.Message, "Invalid Role")
		})
	})
}
