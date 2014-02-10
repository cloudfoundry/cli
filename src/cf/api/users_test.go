package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	ccReqs = []testnet.TestRequest{
		testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   rolePath,
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: fmt.Sprintf(`
				{
					"next_url": "%s",
					"resources": [
						{"metadata": {"guid": "user-1-guid"}, "entity": {}}
					]
				}`, nextUrl)}}),

		testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   nextUrl,
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `
				{
					"resources": [
					 	{"metadata": {"guid": "user-2-guid"}, "entity": {}},
					 	{"metadata": {"guid": "user-3-guid"}, "entity": {}}
					]
				}`}}),
	}

	uaaReqs = []testnet.TestRequest{
		testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path: fmt.Sprintf(
				"/Users?attributes=id,userName&filter=%s",
				url.QueryEscape(`Id eq "user-1-guid" or Id eq "user-2-guid" or Id eq "user-3-guid"`)),
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `
				{
					"resources": [
						{ "id": "user-1-guid", "userName": "Super user 1" },
						{ "id": "user-2-guid", "userName": "Super user 2" },
  						{ "id": "user-3-guid", "userName": "Super user 3" }
					]
				}`}})}

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
	org := models.OrganizationFields{Guid: "some-org-guid"}

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

var _ = Describe("UserRepository", func() {
	Describe("listing the users with a given role", func() {
		It("lists the users in an organization with a given role", func() {
			ccReqs, uaaReqs := createUsersByRoleEndpoints("/v2/organizations/my-org-guid/managers")

			cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(mr.T(), ccReqs, uaaReqs)
			defer cc.Close()
			defer uaa.Close()

			users, apiResponse := repo.ListUsersInOrgForRole("my-org-guid", models.ORG_MANAGER)

			assert.True(mr.T(), ccHandler.AllRequestsCalled())
			assert.True(mr.T(), uaaHandler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.Equal(mr.T(), len(users), 3)
			assert.Equal(mr.T(), users[0].Guid, "user-1-guid")
			assert.Equal(mr.T(), users[0].Username, "Super user 1")
			assert.Equal(mr.T(), users[1].Guid, "user-2-guid")
			assert.Equal(mr.T(), users[1].Username, "Super user 2")
		})

		It("lists the users in a space with a given role", func() {
			ccReqs, uaaReqs := createUsersByRoleEndpoints("/v2/spaces/my-space-guid/managers")

			cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(mr.T(), ccReqs, uaaReqs)
			defer cc.Close()
			defer uaa.Close()

			users, apiResponse := repo.ListUsersInSpaceForRole("my-space-guid", models.SPACE_MANAGER)

			assert.True(mr.T(), ccHandler.AllRequestsCalled())
			assert.True(mr.T(), uaaHandler.AllRequestsCalled())
			assert.True(mr.T(), apiResponse.IsSuccessful())

			assert.Equal(mr.T(), len(users), 3)
			assert.Equal(mr.T(), users[0].Guid, "user-1-guid")
			assert.Equal(mr.T(), users[0].Username, "Super user 1")
			assert.Equal(mr.T(), users[1].Guid, "user-2-guid")
			assert.Equal(mr.T(), users[1].Username, "Super user 2")
		})

		It("does not make a request to the UAA when the cloud controller returns an error", func() {
			ccReqs := []testnet.TestRequest{
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "GET",
					Path:   "/v2/organizations/my-org-guid/managers",
					Response: testnet.TestResponse{
						Status: http.StatusGatewayTimeout,
					},
				}),
			}

			cc, ccHandler, _, _, repo := createUsersRepo(mr.T(), ccReqs, []testnet.TestRequest{})
			defer cc.Close()

			_, apiResponse := repo.ListUsersInOrgForRole("my-org-guid", models.ORG_MANAGER)

			assert.True(mr.T(), ccHandler.AllRequestsCalled())
			assert.Equal(mr.T(), apiResponse.StatusCode, http.StatusGatewayTimeout)
		})

		It("returns an error when the UAA endpoint cannot be determined", func() {
			ccReqs, _ := createUsersByRoleEndpoints("/v2/organizations/my-org-guid/managers")

			cc, _ := testnet.NewTLSServer(mr.T(), ccReqs)
			config := &configuration.Configuration{
				AccessToken:        "BEARER my_access_token",
				Target:             cc.URL,
				OrganizationFields: models.OrganizationFields{Guid: "my-org-guid"},
			}
			ccGateway := net.NewCloudControllerGateway()
			uaaGateway := net.NewUAAGateway()
			endpointRepo := &testapi.FakeEndpointRepo{}
			endpointRepo.UAAEndpointReturns.ApiResponse = net.NewApiResponseWithError("Failed to get endpoint!", errors.New("Failed!"))

			repo := NewCloudControllerUserRepository(config, uaaGateway, ccGateway, endpointRepo)

			_, apiResponse := repo.ListUsersInOrgForRole("my-org-guid", models.ORG_MANAGER)

			Expect(apiResponse).To(Equal(endpointRepo.UAAEndpointReturns.ApiResponse))
		})
	})

	It("TestFindByUsername", func() {
		usersResponse := `{ "resources": [{ "id": "my-guid", "userName": "my-full-username" }]}`
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

		expectedUserFields := models.UserFields{}
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
			Method:   "DELETE",
			Path:     "/v2/users/my-user-guid",
			Response: testnet.TestResponse{Status: http.StatusNotFound, Body: `{"code": 20003, "description": "The user could not be found"}`},
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

	It("lists all users in the org", func() {
		t := mr.T()
		ccReqs, uaaReqs := createUsersByRoleEndpoints("/v2/organizations/my-org-guid/users")

		cc, ccHandler, uaa, uaaHandler, repo := createUsersRepo(t, ccReqs, uaaReqs)
		defer cc.Close()
		defer uaa.Close()

		users, apiResponse := repo.ListUsersInOrgForRole("my-org-guid", models.ORG_USER)

		assert.True(t, ccHandler.AllRequestsCalled())
		assert.True(t, uaaHandler.AllRequestsCalled())
		assert.True(t, apiResponse.IsSuccessful())

		assert.Equal(t, len(users), 3)
		assert.Equal(t, users[0].Guid, "user-1-guid")
		assert.Equal(t, users[0].Username, "Super user 1")
		assert.Equal(t, users[1].Guid, "user-2-guid")
		assert.Equal(t, users[1].Username, "Super user 2")
		assert.Equal(t, users[2].Guid, "user-3-guid")
		assert.Equal(t, users[2].Username, "Super user 3")
	})
})
