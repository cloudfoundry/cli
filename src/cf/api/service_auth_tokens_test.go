package api_test

import (
	. "cf/api"
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
		apiResponse := repo.Create(authToken)

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})
	It("TestServiceAuthFindAll", func() {

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/service_auth_tokens",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `{ "resources": [
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
		]}`},
		})

		ts, handler, repo := createServiceAuthTokenRepo(req)
		defer ts.Close()

		authTokens, apiResponse := repo.FindAll()
		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())

		Expect(len(authTokens)).To(Equal(2))

		Expect(authTokens[0].Label).To(Equal("mysql"))
		Expect(authTokens[0].Provider).To(Equal("mysql-core"))
		Expect(authTokens[0].Guid).To(Equal("mysql-core-guid"))

		Expect(authTokens[1].Label).To(Equal("postgres"))
		Expect(authTokens[1].Provider).To(Equal("postgres-core"))
		Expect(authTokens[1].Guid).To(Equal("postgres-core-guid"))
	})
	It("TestServiceAuthFindByLabelAndProvider", func() {

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/service_auth_tokens?q=label:a-label;provider:a-provider",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body: `{"resources": [{
		"metadata": { "guid": "mysql-core-guid" },
		"entity": {
			"label": "mysql",
			"provider": "mysql-core"
		}
	}]}`},
		})

		ts, handler, repo := createServiceAuthTokenRepo(req)
		defer ts.Close()

		serviceAuthToken, apiResponse := repo.FindByLabelAndProvider("a-label", "a-provider")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
		authToken2 := models.ServiceAuthTokenFields{}
		authToken2.Guid = "mysql-core-guid"
		authToken2.Label = "mysql"
		authToken2.Provider = "mysql-core"
		Expect(serviceAuthToken).To(Equal(authToken2))
	})
	It("TestServiceAuthFindByLabelAndProviderWhenNotFound", func() {

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/service_auth_tokens?q=label:a-label;provider:a-provider",
			Response: testnet.TestResponse{
				Status: http.StatusOK,
				Body:   `{"resources": []}`},
		})

		ts, handler, repo := createServiceAuthTokenRepo(req)
		defer ts.Close()

		_, apiResponse := repo.FindByLabelAndProvider("a-label", "a-provider")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsError()).To(BeFalse())
		Expect(apiResponse.IsNotFound()).To(BeTrue())
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
		apiResponse := repo.Update(authToken3)

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
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
		apiResponse := repo.Delete(authToken4)

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})
})

func createServiceAuthTokenRepo(request testnet.TestRequest) (ts *httptest.Server, handler *testnet.TestHandler, repo ServiceAuthTokenRepository) {
	ts, handler = testnet.NewTLSServer(GinkgoT(), []testnet.TestRequest{request})
	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo = NewCloudControllerServiceAuthTokenRepository(configRepo, gateway)
	return
}
