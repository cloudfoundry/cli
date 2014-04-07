/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package api_test

import (
	. "cf/api"
	"cf/errors"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestServiceAuthCreate", func() {
		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "POST",
			Path:     "/v2/service_auth_tokens",
			Matcher:  testnet.RequestBodyMatcher(`{"label":"a label","provider":"a provider","token":"a token"}`),
			Response: testnet.TestResponse{Status: http.StatusCreated},
		})

		ts, handler, repo := createServiceAuthTokenRepo(req)
		defer ts.Close()
		authToken := models.ServiceAuthTokenFields{}
		authToken.Label = "a label"
		authToken.Provider = "a provider"
		authToken.Token = "a token"
		apiErr := repo.Create(authToken)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
	It("TestServiceAuthFindAll", func() {
		ts, handler, repo := createServiceAuthTokenRepo(serviceAuthRequest)
		defer ts.Close()

		authTokens, apiErr := repo.FindAll()
		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(len(authTokens)).To(Equal(2))

		Expect(authTokens[0].Label).To(Equal("mysql"))
		Expect(authTokens[0].Provider).To(Equal("mysql-core"))
		Expect(authTokens[0].Guid).To(Equal("mysql-core-guid"))

		Expect(authTokens[1].Label).To(Equal("postgres"))
		Expect(authTokens[1].Provider).To(Equal("postgres-core"))
		Expect(authTokens[1].Guid).To(Equal("postgres-core-guid"))
	})

	It("TestServiceAuthFindAll multipage", func() {
		ts, handler, repo := createServiceAuthTokenRepo2([]testnet.TestRequest{firstServiceAuthRequest, serviceAuthRequest})
		defer ts.Close()

		authTokens, apiErr := repo.FindAll()

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())

		Expect(len(authTokens)).To(Equal(3))

		Expect(authTokens[0].Label).To(Equal("mongodb"))
		Expect(authTokens[0].Provider).To(Equal("mongodb-core"))
		Expect(authTokens[0].Guid).To(Equal("mongodb-core-guid"))

		Expect(authTokens[1].Label).To(Equal("mysql"))
		Expect(authTokens[1].Provider).To(Equal("mysql-core"))
		Expect(authTokens[1].Guid).To(Equal("mysql-core-guid"))
	})

	It("TestServiceAuthFindByLabelAndProvider", func() {

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/service_auth_tokens?q=label%3Aa-label%3Bprovider%3Aa-provider",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `{
				"resources": [{
					"metadata": { "guid": "mysql-core-guid" },
					"entity": {
						"label": "mysql",
						"provider": "mysql-core"
					}
				}]}`,
			},
		})

		ts, handler, repo := createServiceAuthTokenRepo(req)
		defer ts.Close()

		serviceAuthToken, apiErr := repo.FindByLabelAndProvider("a-label", "a-provider")

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
		authToken2 := models.ServiceAuthTokenFields{}
		authToken2.Guid = "mysql-core-guid"
		authToken2.Label = "mysql"
		authToken2.Provider = "mysql-core"
		Expect(serviceAuthToken).To(Equal(authToken2))
	})
	It("TestServiceAuthFindByLabelAndProviderWhenNotFound", func() {

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/service_auth_tokens?q=label%3Aa-label%3Bprovider%3Aa-provider",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body:   `{"resources": []}`},
		})

		ts, handler, repo := createServiceAuthTokenRepo(req)
		defer ts.Close()

		_, apiErr := repo.FindByLabelAndProvider("a-label", "a-provider")

		Expect(handler).To(testnet.HaveAllRequestsCalled())

		Expect(apiErr.(*errors.ModelNotFoundError)).NotTo(BeNil())
	})
	It("TestServiceAuthUpdate", func() {

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/v2/service_auth_tokens/mysql-core-guid",
			Matcher:  testnet.RequestBodyMatcher(`{"token":"a value"}`),
			Response: testnet.TestResponse{Status: http.StatusOK},
		})

		ts, handler, repo := createServiceAuthTokenRepo(req)
		defer ts.Close()
		authToken3 := models.ServiceAuthTokenFields{}
		authToken3.Guid = "mysql-core-guid"
		authToken3.Token = "a value"
		apiErr := repo.Update(authToken3)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
	It("TestServiceAuthDelete", func() {

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "DELETE",
			Path:     "/v2/service_auth_tokens/mysql-core-guid",
			Response: testnet.TestResponse{Status: http.StatusOK},
		})

		ts, handler, repo := createServiceAuthTokenRepo(req)
		defer ts.Close()
		authToken4 := models.ServiceAuthTokenFields{}
		authToken4.Guid = "mysql-core-guid"
		apiErr := repo.Delete(authToken4)

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
})

var serviceAuthRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_auth_tokens",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
		"resources": [
			{
				"metadata": {
					"guid": "mysql-core-guid"
			  	},
			  	"entity": {
					"label": "mysql",
					"provider": "mysql-core"
			  	}
			},
			{
				"metadata": {
					"guid": "postgres-core-guid"
			  	},
			  	"entity": {
					"label": "postgres",
					"provider": "postgres-core"
			  	}
			}
		]}`,
	},
})

var firstServiceAuthRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method: "GET",
	Path:   "/v2/service_auth_tokens",
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `{
		"next_url": "/v2/service_auth_tokens?page=2",		    
		"resources": [
			{
				"metadata": {
					"guid": "mongodb-core-guid"
			  	},
			  	"entity": {
					"label": "mongodb",
					"provider": "mongodb-core"
			  	}
			}
		]}`,
	},
})

func createServiceAuthTokenRepo(request testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceAuthTokenRepository) {
	return createServiceAuthTokenRepo2([]testnet.TestRequest{request})
}

func createServiceAuthTokenRepo2(requests []testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceAuthTokenRepository) {
	ts, handler = testnet.NewServer(requests)
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	repo = NewCloudControllerServiceAuthTokenRepository(configRepo, gateway)
	return
}
