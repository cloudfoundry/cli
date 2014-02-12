package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"net/http/httptest"
	"strings"
	testconfig "testhelpers/configuration"
)

var validApiInfoEndpoint = func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v2/info" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	infoResponse := `
{
  "name": "vcap",
  "build": "2222",
  "support": "http://support.cloudfoundry.com",
  "version": 2,
  "description": "Cloud Foundry sponsored by Pivotal",
  "authorization_endpoint": "https://login.example.com",
  "logging_endpoint": "wss://loggregator.foo.example.org:4443",
  "api_version": "42.0.0"
} `
	fmt.Fprintln(w, infoResponse)
}

var ApiInfoEndpointWithoutLogURL = func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v2/info" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, `
{
  "name": "vcap",
  "build": "2222",
  "support": "http://support.cloudfoundry.com",
  "version": 2,
  "description": "Cloud Foundry sponsored by Pivotal",
  "authorization_endpoint": "https://login.example.com",
  "api_version": "42.0.0"
}`)
}

var invalidJsonResponseApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func createEndpointRepoForUpdate(config configuration.ReadWriter, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	if endpoint != nil {
		ts = httptest.NewTLSServer(http.HandlerFunc(endpoint))
	}
	gateway := net.NewCloudControllerGateway()
	return ts, NewEndpointRepository(config, gateway)
}

func createInsecureEndpointRepoForUpdate(config configuration.ReadWriter, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	if endpoint != nil {
		ts = httptest.NewServer(http.HandlerFunc(endpoint))
	}
	gateway := net.NewCloudControllerGateway()
	return ts, NewEndpointRepository(config, gateway)
}

var _ = Describe("Testing with ginkgo", func() {
	var config configuration.ReadWriter

	BeforeEach(func() {
		config = testconfig.NewRepository()
	})

	It("TestUpdateEndpointWhenUrlIsValidHttpsInfoEndpoint", func() {
		ts, repo := createEndpointRepoForUpdate(config, validApiInfoEndpoint)
		defer ts.Close()

		org := models.OrganizationFields{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"

		space := models.SpaceFields{}
		space.Name = "my-space"
		space.Guid = "my-space-guid"

		config.SetOrganizationFields(org)
		config.SetSpaceFields(space)

		repo.UpdateEndpoint(ts.URL)

		assert.Equal(mr.T(), config.AccessToken(), "")
		assert.Equal(mr.T(), config.AuthorizationEndpoint(), "https://login.example.com")
		assert.Equal(mr.T(), config.LoggregatorEndpoint(), "wss://loggregator.foo.example.org:4443")
		assert.Equal(mr.T(), config.ApiEndpoint(), ts.URL)
		assert.Equal(mr.T(), config.ApiVersion(), "42.0.0")
		assert.False(mr.T(), config.HasOrganization())
		assert.False(mr.T(), config.HasSpace())
	})

	It("TestUpdateEndpointWhenUrlIsAlreadyTargeted", func() {
		ts, repo := createEndpointRepoForUpdate(config, validApiInfoEndpoint)
		defer ts.Close()

		org := models.OrganizationFields{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"

		space := models.SpaceFields{}
		space.Name = "my-space"
		space.Guid = "my-space-guid"

		config.SetApiEndpoint(ts.URL)
		config.SetAccessToken("some access token")
		config.SetRefreshToken("some refresh token")
		config.SetOrganizationFields(org)
		config.SetSpaceFields(space)

		repo.UpdateEndpoint(ts.URL)

		assert.Equal(mr.T(), config.OrganizationFields(), org)
		assert.Equal(mr.T(), config.SpaceFields(), space)
		assert.Equal(mr.T(), config.AccessToken(), "some access token")
		assert.Equal(mr.T(), config.RefreshToken(), "some refresh token")
	})

	It("TestUpdateEndpointWhenUrlIsMissingSchemeAndHttpsEndpointExists", func() {
		ts, repo := createEndpointRepoForUpdate(config, validApiInfoEndpoint)
		defer ts.Close()

		schemelessURL := strings.Replace(ts.URL, "https://", "", 1)
		endpoint, apiResponse := repo.UpdateEndpoint(schemelessURL)
		assert.Equal(mr.T(), "https://"+schemelessURL, endpoint)

		assert.True(mr.T(), apiResponse.IsSuccessful())

		assert.Equal(mr.T(), config.AccessToken(), "")
		assert.Equal(mr.T(), config.AuthorizationEndpoint(), "https://login.example.com")
		assert.Equal(mr.T(), config.ApiEndpoint(), ts.URL)
		assert.Equal(mr.T(), config.ApiVersion(), "42.0.0")
	})

	It("TestUpdateEndpointWhenUrlIsMissingSchemeAndHttpEndpointExists", func() {
		ts, repo := createInsecureEndpointRepoForUpdate(config, validApiInfoEndpoint)
		defer ts.Close()

		schemelessURL := strings.Replace(ts.URL, "http://", "", 1)

		endpoint, apiResponse := repo.UpdateEndpoint(schemelessURL)
		assert.Equal(mr.T(), "http://"+schemelessURL, endpoint)

		assert.True(mr.T(), apiResponse.IsSuccessful())

		assert.Equal(mr.T(), config.AccessToken(), "")
		assert.Equal(mr.T(), config.AuthorizationEndpoint(), "https://login.example.com")
		assert.Equal(mr.T(), config.ApiEndpoint(), ts.URL)
		assert.Equal(mr.T(), config.ApiVersion(), "42.0.0")
	})

	It("TestUpdateEndpointWhenEndpointReturns404", func() {
		ts, repo := createEndpointRepoForUpdate(config, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		defer ts.Close()

		_, apiResponse := repo.UpdateEndpoint(ts.URL)

		assert.True(mr.T(), apiResponse.IsNotSuccessful())
	})

	It("TestUpdateEndpointWhenEndpointReturnsInvalidJson", func() {
		ts, repo := createEndpointRepoForUpdate(config, invalidJsonResponseApiEndpoint)
		defer ts.Close()

		_, apiResponse := repo.UpdateEndpoint(ts.URL)

		assert.True(mr.T(), apiResponse.IsNotSuccessful())
	})

	It("TestGetCloudControllerEndpoint", func() {
		config.SetApiEndpoint("http://api.example.com")

		repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

		endpoint, apiResponse := repo.GetCloudControllerEndpoint()

		assert.True(mr.T(), apiResponse.IsSuccessful())
		assert.Equal(mr.T(), endpoint, "http://api.example.com")
	})

	It("TestGetLoggregatorEndpoint", func() {
		config.SetLoggregatorEndpoint("wss://loggregator.example.com:4443")

		repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

		endpoint, apiResponse := repo.GetLoggregatorEndpoint()

		assert.True(mr.T(), apiResponse.IsSuccessful())
		assert.Equal(mr.T(), endpoint, "wss://loggregator.example.com:4443")
	})

	Describe("when the loggregator endpoint is not saved in the config (old CC)", func() {
		BeforeEach(func() {
			config.SetLoggregatorEndpoint("")
		})

		It("extrapolates the loggregator URL based on the API URL (SSL API)", func() {
			config.SetApiEndpoint("https://api.run.pivotal.io")

			repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

			endpoint, apiResponse := repo.GetLoggregatorEndpoint()
			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), endpoint, "wss://loggregator.run.pivotal.io:4443")
		})

		It("extrapolates the loggregator URL based on the API URL (non-SSL API)", func() {
			config.SetApiEndpoint("http://api.run.pivotal.io")

			repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

			endpoint, apiResponse := repo.GetLoggregatorEndpoint()
			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), endpoint, "ws://loggregator.run.pivotal.io:80")
		})
	})

	It("TestGetUAAEndpoint", func() {
		config := testconfig.NewRepository()
		config.SetAuthorizationEndpoint("https://login.example.com")

		repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

		endpoint, apiResponse := repo.GetUAAEndpoint()

		assert.True(mr.T(), apiResponse.IsSuccessful())
		assert.Equal(mr.T(), endpoint, "https://uaa.example.com")
	})

	It("TestEndpointsReturnAnErrorWhenMissing", func() {
		config := testconfig.NewRepository()
		repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

		_, response := repo.GetLoggregatorEndpoint()
		assert.True(mr.T(), response.IsNotSuccessful())

		_, response = repo.GetCloudControllerEndpoint()
		assert.True(mr.T(), response.IsNotSuccessful())

		_, response = repo.GetUAAEndpoint()
		assert.True(mr.T(), response.IsNotSuccessful())
	})
})
