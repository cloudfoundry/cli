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

var invalidJsonResponseApiEndpoint = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `Foo`)
}

func createEndpointRepoForUpdate(configRepo testconfig.FakeConfigRepository, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	if endpoint != nil {
		ts = httptest.NewTLSServer(http.HandlerFunc(endpoint))
	}
	return ts, makeRepo(configRepo)
}

func createInsecureEndpointRepoForUpdate(configRepo testconfig.FakeConfigRepository, endpoint func(w http.ResponseWriter, r *http.Request)) (ts *httptest.Server, repo EndpointRepository) {
	if endpoint != nil {
		ts = httptest.NewServer(http.HandlerFunc(endpoint))
	}
	return ts, makeRepo(configRepo)
}

func makeRepo(configRepo testconfig.FakeConfigRepository) (repo EndpointRepository) {
	config, _ := configRepo.Get()
	gateway := net.NewCloudControllerGateway()
	return NewEndpointRepository(config, gateway, configRepo)
}

func createEndpointRepoForGet(config *configuration.Configuration) (repo EndpointRepository) {
	configRepo := testconfig.FakeConfigRepository{}
	repo = NewEndpointRepository(config, net.NewCloudControllerGateway(), configRepo)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUpdateEndpointWhenUrlIsValidHttpsInfoEndpoint", func() {
			configRepo := testconfig.FakeConfigRepository{}
			configRepo.Delete()
			configRepo.Login()

			ts, repo := createEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
			defer ts.Close()
			org := cf.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			space := cf.SpaceFields{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"

			config, _ := configRepo.Get()
			config.OrganizationFields = org
			config.SpaceFields = space

			repo.UpdateEndpoint(ts.URL)

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.AccessToken, "")
			assert.Equal(mr.T(), savedConfig.AuthorizationEndpoint, "https://login.example.com")
			assert.Equal(mr.T(), savedConfig.LoggregatorEndPoint, "wss://loggregator.foo.example.org:4443")
			assert.Equal(mr.T(), savedConfig.Target, ts.URL)
			assert.Equal(mr.T(), savedConfig.ApiVersion, "42.0.0")
			assert.False(mr.T(), savedConfig.HasOrganization())
			assert.False(mr.T(), savedConfig.HasSpace())
		})
		It("TestUpdateEndpointWhenUrlIsAlreadyTargeted", func() {

			configRepo := testconfig.FakeConfigRepository{}
			configRepo.Delete()
			configRepo.Login()

			ts, repo := createEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
			defer ts.Close()

			org := cf.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			space := cf.SpaceFields{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"

			config, _ := configRepo.Get()
			config.Target = ts.URL
			config.AccessToken = "some access token"
			config.RefreshToken = "some refresh token"
			config.OrganizationFields = org
			config.SpaceFields = space

			repo.UpdateEndpoint(ts.URL)

			assert.Equal(mr.T(), config.OrganizationFields, org)
			assert.Equal(mr.T(), config.SpaceFields, space)
			assert.Equal(mr.T(), config.AccessToken, "some access token")
			assert.Equal(mr.T(), config.RefreshToken, "some refresh token")
		})
		It("TestUpdateEndpointWhenUrlIsMissingSchemeAndHttpsEndpointExists", func() {

			configRepo := testconfig.FakeConfigRepository{}
			configRepo.Delete()
			configRepo.Login()

			ts, repo := createEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
			defer ts.Close()

			schemelessURL := strings.Replace(ts.URL, "https://", "", 1)
			endpoint, apiResponse := repo.UpdateEndpoint(schemelessURL)
			assert.Equal(mr.T(), "https://"+schemelessURL, endpoint)

			assert.True(mr.T(), apiResponse.IsSuccessful())

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.AccessToken, "")
			assert.Equal(mr.T(), savedConfig.AuthorizationEndpoint, "https://login.example.com")
			assert.Equal(mr.T(), savedConfig.Target, ts.URL)
			assert.Equal(mr.T(), savedConfig.ApiVersion, "42.0.0")
		})
		It("TestUpdateEndpointWhenUrlIsMissingSchemeAndHttpEndpointExists", func() {

			configRepo := testconfig.FakeConfigRepository{}
			configRepo.Delete()
			configRepo.Login()

			ts, repo := createInsecureEndpointRepoForUpdate(configRepo, validApiInfoEndpoint)
			defer ts.Close()

			schemelessURL := strings.Replace(ts.URL, "http://", "", 1)

			endpoint, apiResponse := repo.UpdateEndpoint(schemelessURL)
			assert.Equal(mr.T(), "http://"+schemelessURL, endpoint)

			assert.True(mr.T(), apiResponse.IsSuccessful())

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.AccessToken, "")
			assert.Equal(mr.T(), savedConfig.AuthorizationEndpoint, "https://login.example.com")
			assert.Equal(mr.T(), savedConfig.Target, ts.URL)
			assert.Equal(mr.T(), savedConfig.ApiVersion, "42.0.0")
		})
		It("TestUpdateEndpointWhenEndpointReturns404", func() {

			configRepo := testconfig.FakeConfigRepository{}
			configRepo.Login()

			ts, repo := createEndpointRepoForUpdate(configRepo, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})

			defer ts.Close()

			_, apiResponse := repo.UpdateEndpoint(ts.URL)

			assert.True(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestUpdateEndpointWhenEndpointReturnsInvalidJson", func() {

			configRepo := testconfig.FakeConfigRepository{}
			configRepo.Login()

			ts, repo := createEndpointRepoForUpdate(configRepo, invalidJsonResponseApiEndpoint)
			defer ts.Close()

			_, apiResponse := repo.UpdateEndpoint(ts.URL)

			assert.True(mr.T(), apiResponse.IsNotSuccessful())
		})
		It("TestGetCloudControllerEndpoint", func() {

			configRepo := testconfig.FakeConfigRepository{}
			config := &configuration.Configuration{
				Target: "http://api.example.com",
			}

			repo := NewEndpointRepository(config, net.NewCloudControllerGateway(), configRepo)

			endpoint, apiResponse := repo.GetCloudControllerEndpoint()

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), endpoint, "http://api.example.com")
		})
		It("TestGetLoggregatorEndpoint", func() {

			config := &configuration.Configuration{
				LoggregatorEndPoint: "wss://loggregator.example.com:4443",
			}

			repo := createEndpointRepoForGet(config)

			endpoint, apiResponse := repo.GetLoggregatorEndpoint()

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), endpoint, "wss://loggregator.example.com:4443")
		})
		It("TestGetUAAEndpoint", func() {

			config := &configuration.Configuration{
				AuthorizationEndpoint: "https://login.example.com",
			}

			repo := createEndpointRepoForGet(config)

			endpoint, apiResponse := repo.GetUAAEndpoint()

			assert.True(mr.T(), apiResponse.IsSuccessful())
			assert.Equal(mr.T(), endpoint, "https://uaa.example.com")
		})
		It("TestEndpointsReturnAnErrorWhenMissing", func() {

			config := &configuration.Configuration{}
			repo := createEndpointRepoForGet(config)

			_, response := repo.GetLoggregatorEndpoint()
			assert.True(mr.T(), response.IsNotSuccessful())

			_, response = repo.GetCloudControllerEndpoint()
			assert.True(mr.T(), response.IsNotSuccessful())

			_, response = repo.GetUAAEndpoint()
			assert.True(mr.T(), response.IsNotSuccessful())
		})
	})
}
