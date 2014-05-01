package api_test

import (
	"crypto/tls"
	"fmt"
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"
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

func apiInfoEndpointWithoutLogURL(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/v2/info") {
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
		gateway = net.NewCloudControllerGateway(config)
		gateway.SetTrustedCerts(testServer.TLS.Certificates)
		repo = NewEndpointRepository(config, gateway)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("updating the endpoints", func() {
		Context("when the API request is successful", func() {
			var (
				org   models.OrganizationFields
				space models.SpaceFields
			)
			BeforeEach(func() {
				org.Name = "my-org"
				org.Guid = "my-org-guid"

				space.Name = "my-space"
				space.Guid = "my-space-guid"

				config.SetOrganizationFields(org)
				config.SetSpaceFields(space)
				testServerFn = validApiInfoEndpoint
			})

			It("stores the data from the /info api in the config", func() {
				repo.UpdateEndpoint(testServer.URL)

				Expect(config.AccessToken()).To(Equal(""))
				Expect(config.AuthenticationEndpoint()).To(Equal("https://login.example.com"))
				Expect(config.LoggregatorEndpoint()).To(Equal("wss://loggregator.foo.example.org:4443"))
				Expect(config.ApiEndpoint()).To(Equal(testServer.URL))
				Expect(config.ApiVersion()).To(Equal("42.0.0"))
				Expect(config.HasOrganization()).To(BeFalse())
				Expect(config.HasSpace()).To(BeFalse())
			})

			Context("when the api endpoint does not change", func() {
				BeforeEach(func() {
					config.SetApiEndpoint(testServer.URL)
					config.SetAccessToken("some access token")
					config.SetRefreshToken("some refresh token")
				})

				It("does not clear the session if the api endpoint does not change", func() {
					repo.UpdateEndpoint(testServer.URL)

					Expect(config.OrganizationFields()).To(Equal(org))
					Expect(config.SpaceFields()).To(Equal(space))
					Expect(config.AccessToken()).To(Equal("some access token"))
					Expect(config.RefreshToken()).To(Equal("some refresh token"))
				})
			})
		})

		Context("when the API request fails", func() {
			ItClearsTheConfig := func() {
				Expect(config.ApiEndpoint()).To(BeEmpty())
			}

			BeforeEach(func() {
				config.SetApiEndpoint("example.com")
			})

			It("returns a failure response when the server has a bad certificate", func() {
				testServer.TLS.Certificates = []tls.Certificate{testnet.MakeExpiredTLSCert()}

				_, apiErr := repo.UpdateEndpoint(testServer.URL)
				Expect(apiErr).NotTo(BeNil())
				ItClearsTheConfig()
			})

			It("returns a failure response when the API request fails", func() {
				testServerFn = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}

				_, apiErr := repo.UpdateEndpoint(testServer.URL)

				Expect(apiErr).NotTo(BeNil())
				ItClearsTheConfig()
			})

			It("returns a failure response when the API returns invalid JSON", func() {
				testServerFn = func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, `Foo`)
				}

				_, apiErr := repo.UpdateEndpoint(testServer.URL)

				Expect(apiErr).NotTo(BeNil())
				ItClearsTheConfig()
			})
		})

		Describe("when the specified API url doesn't have a scheme", func() {
			It("uses https if possible", func() {
				testServerFn = validApiInfoEndpoint

				schemelessURL := strings.Replace(testServer.URL, "https://", "", 1)
				endpoint, apiErr := repo.UpdateEndpoint(schemelessURL)
				Expect(endpoint).To(Equal("https://" + schemelessURL))

				Expect(apiErr).NotTo(HaveOccurred())

				Expect(config.AccessToken()).To(Equal(""))
				Expect(config.AuthenticationEndpoint()).To(Equal("https://login.example.com"))
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
				Expect(config.AuthenticationEndpoint()).To(Equal("https://login.example.com"))
				Expect(config.ApiEndpoint()).To(Equal(testServer.URL))
				Expect(config.ApiVersion()).To(Equal("42.0.0"))
			})
		})

		Describe("when the loggregator endpoint is not returned by the /info API (old CC)", func() {
			BeforeEach(func() {
				testServerFn = apiInfoEndpointWithoutLogURL
			})

			It("extrapolates the loggregator URL based on the API URL (SSL API)", func() {
				_, err := repo.UpdateEndpoint(testServer.URL)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.LoggregatorEndpoint()).To(Equal("wss://loggregator.0.0.1:443"))
			})

			It("extrapolates the loggregator URL if there is a trailing slash", func() {
				_, err := repo.UpdateEndpoint(testServer.URL + "/")
				Expect(err).NotTo(HaveOccurred())
				Expect(config.LoggregatorEndpoint()).To(Equal("wss://loggregator.0.0.1:443"))
			})

			It("extrapolates the loggregator URL based on the API URL (non-SSL API)", func() {
				testServer.Close()
				testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					testServerFn(w, r)
				}))

				_, err := repo.UpdateEndpoint(testServer.URL)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.LoggregatorEndpoint()).To(Equal("ws://loggregator.0.0.1:80"))
			})
		})
	})
})
