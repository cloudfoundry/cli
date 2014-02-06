package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
)

func createOrganizationRepo(t mr.TestingT, reqs ...testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo OrganizationRepository) {
	ts, handler = testnet.NewTLSServer(t, reqs)

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerOrganizationRepository(config, gateway)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestOrganizationsListOrgs", func() {
			firstPageOrgsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{
		"next_url": "/v2/organizations?page=2",
		"resources": [
			{
			  "metadata": { "guid": "org1-guid" },
			  "entity": { "name": "Org1" }
			},
			{
			  "metadata": { "guid": "org2-guid" },
			  "entity": { "name": "Org2" }
			}
		]}`},
			})

			secondPageOrgsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations?page=2",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [
			{
			  "metadata": { "guid": "org3-guid" },
			  "entity": { "name": "Org3" }
			}
		]}`},
			})

			ts, handler, repo := createOrganizationRepo(mr.T(), firstPageOrgsRequest, secondPageOrgsRequest)
			defer ts.Close()

			stopChan := make(chan bool)
			defer close(stopChan)
			orgsChan, statusChan := repo.ListOrgs(stopChan)

			orgs := []models.Organization{}
			for chunk := range orgsChan {
				orgs = append(orgs, chunk...)
			}
			apiResponse := <-statusChan

			assert.Equal(mr.T(), len(orgs), 3)
			assert.Equal(mr.T(), orgs[0].Guid, "org1-guid")
			assert.Equal(mr.T(), orgs[1].Guid, "org2-guid")
			assert.Equal(mr.T(), orgs[2].Guid, "org3-guid")
			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.True(mr.T(), handler.AllRequestsCalled())
		})
		It("TestOrganizationsListOrgsWithNoOrgs", func() {

			emptyOrgsRequest := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/organizations",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
			})

			ts, handler, repo := createOrganizationRepo(mr.T(), emptyOrgsRequest)
			defer ts.Close()

			stopChan := make(chan bool)
			defer close(stopChan)
			orgsChan, statusChan := repo.ListOrgs(stopChan)

			_, ok := <-orgsChan
			apiResponse := <-statusChan

			assert.False(mr.T(), ok)
			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.True(mr.T(), handler.AllRequestsCalled())
		})
		It("TestOrganizationsFindByName", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method: "GET",
				Path:   "/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": [{
		  "metadata": { "guid": "org1-guid" },
		  "entity": {
			"name": "Org1",
			"quota_definition": {
			  "entity": {
			  	"name": "not-your-average-quota",
			    "memory_limit": 128
			  }
			},
			"spaces": [{
			  "metadata": { "guid": "space1-guid" },
			  "entity": { "name": "Space1" }
			}],
			"domains": [{
			  "metadata": { "guid": "domain1-guid" },
			  "entity": { "name": "cfapps.io" }
			}]
		  }
		}]}`},
			})

			ts, handler, repo := createOrganizationRepo(mr.T(), req)
			defer ts.Close()
			existingOrg := models.Organization{}
			existingOrg.Guid = "org1-guid"
			existingOrg.Name = "Org1"

			org, apiResponse := repo.FindByName("Org1")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())

			assert.Equal(mr.T(), org.Name, existingOrg.Name)
			assert.Equal(mr.T(), org.Guid, existingOrg.Guid)
			assert.Equal(mr.T(), org.QuotaDefinition.Name, "not-your-average-quota")
			assert.Equal(mr.T(), org.QuotaDefinition.MemoryLimit, uint64(128))
			assert.Equal(mr.T(), len(org.Spaces), 1)
			assert.Equal(mr.T(), org.Spaces[0].Name, "Space1")
			assert.Equal(mr.T(), org.Spaces[0].Guid, "space1-guid")
			assert.Equal(mr.T(), len(org.Domains), 1)
			assert.Equal(mr.T(), org.Domains[0].Name, "cfapps.io")
			assert.Equal(mr.T(), org.Domains[0].Guid, "domain1-guid")
		})
		It("TestOrganizationsFindByNameWhenDoesNotExist", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "GET",
				Path:     "/v2/organizations?q=name%3Aorg1&inline-relations-depth=1",
				Response: testnet.TestResponse{Status: http.StatusOK, Body: `{"resources": []}`},
			})

			ts, handler, repo := createOrganizationRepo(mr.T(), req)
			defer ts.Close()

			_, apiResponse := repo.FindByName("org1")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsError())
			assert.True(mr.T(), apiResponse.IsNotFound())
		})
		It("TestCreateOrganization", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "POST",
				Path:     "/v2/organizations",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-org"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createOrganizationRepo(mr.T(), req)
			defer ts.Close()

			apiResponse := repo.Create("my-org")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestRenameOrganization", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "PUT",
				Path:     "/v2/organizations/my-org-guid",
				Matcher:  testnet.RequestBodyMatcher(`{"name":"my-new-org"}`),
				Response: testnet.TestResponse{Status: http.StatusCreated},
			})

			ts, handler, repo := createOrganizationRepo(mr.T(), req)
			defer ts.Close()

			apiResponse := repo.Rename("my-org-guid", "my-new-org")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestDeleteOrganization", func() {

			req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:   "DELETE",
				Path:     "/v2/organizations/my-org-guid?recursive=true",
				Response: testnet.TestResponse{Status: http.StatusOK},
			})

			ts, handler, repo := createOrganizationRepo(mr.T(), req)
			defer ts.Close()

			apiResponse := repo.Delete("my-org-guid")
			assert.True(mr.T(), handler.AllRequestsCalled())
			assert.False(mr.T(), apiResponse.IsNotSuccessful())
		})
	})
}
