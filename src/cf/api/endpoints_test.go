package api_test

import (
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"
	testconfig "testhelpers/configuration"
)

func validApiInfoEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v2/info" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, `
{
  "name": "vcap",
  "build": "2222",
  "support": "http://support.cloudfoundry.com",
  "version": 2,
  "description": "Cloud Foundry sponsored by Pivotal",
  "authorization_endpoint": "https://login.example.com",
  "logging_endpoint": "wss://loggregator.foo.example.org:4443",
  "api_version": "42.0.0"
}`)
}

func ApiInfoEndpointWithoutLogURL(w http.ResponseWriter, r *http.Request) {
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

var _ = Describe("Endpoints Repository", func() {
	var (
		config       configuration.ReadWriter
		gateway      net.Gateway
		testServer   *httptest.Server
		repo         EndpointRepository
		testServerFn func(w http.ResponseWriter, r *http.Request)
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
		testServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			testServerFn(w, r)
		}))
		gateway = net.NewCloudControllerGateway()
		gateway.AddTrustedCerts(testServer.TLS.Certificates)
		repo = NewEndpointRepository(config, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("updating the endpoints", func() {
		It("stores the data from the /info api in the config", func() {
			testServerFn = validApiInfoEndpoint

			org := models.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			space := models.SpaceFields{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"

			config.SetOrganizationFields(org)
			config.SetSpaceFields(space)

			repo.UpdateEndpoint(testServer.URL)

			Expect(config.AccessToken()).To(Equal(""))
			Expect(config.AuthorizationEndpoint()).To(Equal("https://login.example.com"))
			Expect(config.LoggregatorEndpoint()).To(Equal("wss://loggregator.foo.example.org:4443"))
			Expect(config.ApiEndpoint()).To(Equal(testServer.URL))
			Expect(config.ApiVersion()).To(Equal("42.0.0"))
			Expect(config.HasOrganization()).To(BeFalse())
			Expect(config.HasSpace()).To(BeFalse())
		})

		It("TestUpdateEndpointWhenUrlIsAlreadyTargeted", func() {
			testServerFn = validApiInfoEndpoint

			org := models.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			space := models.SpaceFields{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"

			config.SetApiEndpoint(testServer.URL)
			config.SetAccessToken("some access token")
			config.SetRefreshToken("some refresh token")
			config.SetOrganizationFields(org)
			config.SetSpaceFields(space)

			repo.UpdateEndpoint(testServer.URL)

			Expect(config.OrganizationFields()).To(Equal(org))
			Expect(config.SpaceFields()).To(Equal(space))
			Expect(config.AccessToken()).To(Equal("some access token"))
			Expect(config.RefreshToken()).To(Equal("some refresh token"))
		})

		It("returns a failure response when the API request fails", func() {
			testServerFn = func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}

			_, apiErr := repo.UpdateEndpoint(testServer.URL)

			Expect(apiErr).NotTo(BeNil())
		})

		It("returns a failure response when the API returns invalid JSON", func() {
			testServerFn = invalidJsonResponseApiEndpoint

			_, apiErr := repo.UpdateEndpoint(testServer.URL)

			Expect(apiErr).NotTo(BeNil())
		})

		Describe("when the specified API url doesn't have a scheme", func() {
			It("uses https if possible", func() {
				testServerFn = validApiInfoEndpoint

				schemelessURL := strings.Replace(testServer.URL, "https://", "", 1)
				endpoint, apiErr := repo.UpdateEndpoint(schemelessURL)
				Expect(endpoint).To(Equal("https://" + schemelessURL))

				Expect(apiErr).NotTo(HaveOccurred())

				Expect(config.AccessToken()).To(Equal(""))
				Expect(config.AuthorizationEndpoint()).To(Equal("https://login.example.com"))
				Expect(config.ApiEndpoint()).To(Equal(testServer.URL))
				Expect(config.ApiVersion()).To(Equal("42.0.0"))
			})

			It("uses http when the server doesn't respond over https", func() {
				testServer.Close()
				testServer = httptest.NewServer(http.HandlerFunc(validApiInfoEndpoint))
				schemelessURL := strings.Replace(testServer.URL, "http://", "", 1)

				endpoint, apiErr := repo.UpdateEndpoint(schemelessURL)

				Expect(endpoint).To(Equal("http://" + schemelessURL))
				Expect(apiErr).NotTo(HaveOccurred())

				Expect(config.AccessToken()).To(Equal(""))
				Expect(config.AuthorizationEndpoint()).To(Equal("https://login.example.com"))
				Expect(config.ApiEndpoint()).To(Equal(testServer.URL))
				Expect(config.ApiVersion()).To(Equal("42.0.0"))
			})
		})
	})

	Describe("getting API endpoints from a saved config", func() {
		It("TestGetCloudControllerEndpoint", func() {
			config.SetApiEndpoint("http://api.example.com")

			repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

			endpoint, apiErr := repo.GetCloudControllerEndpoint()

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(endpoint).To(Equal("http://api.example.com"))
		})

		It("TestGetLoggregatorEndpoint", func() {
			config.SetLoggregatorEndpoint("wss://loggregator.example.com:4443")

			repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

			endpoint, apiErr := repo.GetLoggregatorEndpoint()

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(endpoint).To(Equal("wss://loggregator.example.com:4443"))
		})

		Describe("when the loggregator endpoint is not saved in the config (old CC)", func() {
			BeforeEach(func() {
				config.SetLoggregatorEndpoint("")
			})

			It("extrapolates the loggregator URL based on the API URL (SSL API)", func() {
				config.SetApiEndpoint("https://api.run.pivotal.io")

				repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

				endpoint, apiErr := repo.GetLoggregatorEndpoint()
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(endpoint).To(Equal("wss://loggregator.run.pivotal.io:4443"))
			})

			It("extrapolates the loggregator URL based on the API URL (non-SSL API)", func() {
				config.SetApiEndpoint("http://api.run.pivotal.io")

				repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

				endpoint, apiErr := repo.GetLoggregatorEndpoint()
				Expect(apiErr).NotTo(HaveOccurred())
				Expect(endpoint).To(Equal("ws://loggregator.run.pivotal.io:80"))
			})
		})

		It("TestGetUAAEndpoint", func() {
			config := testconfig.NewRepository()
			config.SetAuthorizationEndpoint("https://login.example.com")

			repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

			endpoint, apiErr := repo.GetUAAEndpoint()

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(endpoint).To(Equal("https://uaa.example.com"))
		})

		It("TestEndpointsReturnAnErrorWhenMissing", func() {
			config := testconfig.NewRepository()
			repo := NewEndpointRepository(config, net.NewCloudControllerGateway())

			_, apiErr := repo.GetLoggregatorEndpoint()
			Expect(apiErr).To(HaveOccurred())

			_, apiErr = repo.GetCloudControllerEndpoint()
			Expect(apiErr).To(HaveOccurred())

			_, apiErr = repo.GetUAAEndpoint()
			Expect(apiErr).To(HaveOccurred())
		})
	})
})
