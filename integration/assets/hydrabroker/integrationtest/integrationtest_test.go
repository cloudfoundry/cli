package integrationtest_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	uuid2 "github.com/nu7hatch/gouuid"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/app"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
)

var _ = Describe("Integration Test", func() {
	var (
		server *httptest.Server
		client *http.Client
	)

	BeforeEach(func() {
		server = httptest.NewServer(app.App())
		client = server.Client()
	})

	AfterEach(func() {
		server.Close()
	})

	It("responds to an aliveness test", func() {
		response, err := client.Head(server.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusNoContent))
	})

	It("allows a broker to be created", func() {
		response, err := client.Post(server.URL+"/config", "application/json", toJSON(randomConfiguration()))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusCreated))

		var r config.NewBrokerResponse
		fromJSON(response.Body, &r)
		Expect(r.GUID).NotTo(BeEmpty())
	})

	When("the create request is missing parameters", func() {
		It("fails", func() {
			response, err := client.Post(server.URL+"/config", "application/json", strings.NewReader("{}"))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))

			b, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(ContainSubstring("Error:Field validation for 'Username' failed on the 'min' tag"))
		})
	})

	When("a broker exists", func() {
		var (
			guid string
			cfg  config.BrokerConfiguration
		)

		BeforeEach(func() {
			cfg = randomConfiguration()
			response, err := client.Post(server.URL+"/config", "application/json", toJSON(cfg))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusCreated))

			var r config.NewBrokerResponse
			fromJSON(response.Body, &r)
			guid = r.GUID
		})

		It("lists the broker", func() {
			response, err := client.Get(server.URL + "/config")
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			var r []string
			fromJSON(response.Body, &r)
			Expect(r).To(HaveLen(1))
			Expect(r[0]).To(Equal(guid))
		})

		It("rejects requests without a password", func() {
			response, err := client.Get(server.URL + "/broker/" + guid + "/v2/catalog")
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		It("responds to the catalog endpoint", func() {
			request, err := http.NewRequest("GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
			Expect(err).NotTo(HaveOccurred())
			request.SetBasicAuth(cfg.Username, cfg.Password)
			response, err := client.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			var catalog apiresponses.CatalogResponse
			fromJSON(response.Body, &catalog)
			Expect(catalog.Services[0].ID).To(Equal(cfg.Services[0].ID))
			Expect(catalog.Services[0].Name).To(Equal(cfg.Services[0].Name))
			Expect(catalog.Services[0].Description).To(Equal(cfg.Services[0].Description))
			Expect(catalog.Services[0].Plans[0].ID).To(Equal(cfg.Services[0].Plans[0].ID))
			Expect(catalog.Services[0].Plans[0].Name).To(Equal(cfg.Services[0].Plans[0].Name))
			Expect(catalog.Services[0].Plans[0].Description).To(Equal(cfg.Services[0].Plans[0].Description))
		})

		It("allows a service instance to be created", func() {
			instanceGUID := randomString()
			request, err := http.NewRequest("PUT", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, nil)
			Expect(err).NotTo(HaveOccurred())
			request.SetBasicAuth(cfg.Username, cfg.Password)
			response, err := client.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusCreated))

			var r map[string]interface{}
			fromJSON(response.Body, &r)
			Expect(r).To(BeEmpty())
		})

		It("allows a service instance to be deleted", func() {
			instanceGUID := randomString()
			request, err := http.NewRequest("DELETE", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, nil)
			Expect(err).NotTo(HaveOccurred())
			request.SetBasicAuth(cfg.Username, cfg.Password)
			response, err := client.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			var r map[string]interface{}
			fromJSON(response.Body, &r)
			Expect(r).To(BeEmpty())
		})

		It("allows the broker to be deleted", func() {
			By("accepting the delete request", func() {
				request, err := http.NewRequest("DELETE", server.URL+"/config/"+guid, nil)
				Expect(err).NotTo(HaveOccurred())
				response, err := client.Do(request)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusNoContent))
			})

			By("no longer responding to the catalog endpoint", func() {
				response, err := client.Get(server.URL + "/broker/" + guid + "/v2/catalog")
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusNotFound))
			})

			By("no longer listing the broker", func() {
				response, err := client.Get(server.URL + "/config")
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				var r []string
				fromJSON(response.Body, &r)
				Expect(r).To(HaveLen(0))
			})
		})

		It("allows the broker to be reconfigured", func() {
			newCfg := randomConfiguration()

			By("accepting the reconfigure request", func() {
				request, err := http.NewRequest("PUT", server.URL+"/config/"+guid, toJSON(newCfg))
				Expect(err).NotTo(HaveOccurred())
				response, err := client.Do(request)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusNoContent))
			})

			By("updating the catalog", func() {
				request, err := http.NewRequest("GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
				Expect(err).NotTo(HaveOccurred())
				request.SetBasicAuth(newCfg.Username, newCfg.Password)
				response, err := client.Do(request)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				var catalog apiresponses.CatalogResponse
				fromJSON(response.Body, &catalog)
				Expect(catalog.Services[0].Name).To(Equal(newCfg.Services[0].Name))
			})
		})
	})

	Describe("custom responses", func() {
		var (
			guid string
			cfg  config.BrokerConfiguration
		)

		BeforeEach(func() {
			cfg = randomConfiguration()

			cfg.CatalogResponse = http.StatusInternalServerError
			cfg.ProvisionResponse = http.StatusBadGateway
			cfg.DeprovisionResponse = http.StatusTeapot

			response, err := client.Post(server.URL+"/config", "application/json", toJSON(cfg))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusCreated))

			var r config.NewBrokerResponse
			fromJSON(response.Body, &r)
			guid = r.GUID
		})

		It("allows custom catalog responses", func() {
			request, err := http.NewRequest("GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
			Expect(err).NotTo(HaveOccurred())
			request.SetBasicAuth(cfg.Username, cfg.Password)
			response, err := client.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusInternalServerError))
		})

		It("allows custom provision responses", func() {
			instanceGUID := randomString()
			request, err := http.NewRequest("PUT", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, nil)
			Expect(err).NotTo(HaveOccurred())
			request.SetBasicAuth(cfg.Username, cfg.Password)
			response, err := client.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadGateway))
		})

		It("allows custom deprovision response", func() {
			instanceGUID := randomString()
			request, err := http.NewRequest("DELETE", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, nil)
			Expect(err).NotTo(HaveOccurred())
			request.SetBasicAuth(cfg.Username, cfg.Password)
			response, err := client.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusTeapot))
		})
	})

	Describe("configuring the catalog", func() {
		It("can be configured with maintenance info", func() {
			var guid string
			cfg := randomConfiguration()

			By("accepting the configuration", func() {
				cfg.Services[0].Plans[0].MaintenanceInfo = &config.MaintenanceInfo{
					Version:     "1.2.3",
					Description: "a description",
				}

				response, err := client.Post(server.URL+"/config", "application/json", toJSON(cfg))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusCreated))

				var r config.NewBrokerResponse
				fromJSON(response.Body, &r)
				guid = r.GUID
			})

			By("showing it in the catalog", func() {
				request, err := http.NewRequest("GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
				Expect(err).NotTo(HaveOccurred())
				request.SetBasicAuth(cfg.Username, cfg.Password)
				response, err := client.Do(request)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				var catalog apiresponses.CatalogResponse
				fromJSON(response.Body, &catalog)
				Expect(catalog.Services[0].Name).To(Equal(cfg.Services[0].Name))
				Expect(catalog.Services[0].Plans[0].MaintenanceInfo.Version).To(Equal("1.2.3"))
				Expect(catalog.Services[0].Plans[0].MaintenanceInfo.Description).To(Equal("a description"))
			})
		})
	})
})

func randomConfiguration() config.BrokerConfiguration {
	return config.BrokerConfiguration{
		Services: []config.Service{
			{
				Name:        randomString(),
				ID:          randomString(),
				Description: randomString(),
				Plans: []config.Plan{
					{
						Name:        randomString(),
						ID:          randomString(),
						Description: randomString(),
					},
				},
			},
		},
		Username: randomString(),
		Password: randomString(),
	}
}

func randomString() string {
	uuid, err := uuid2.NewV4()
	Expect(err).NotTo(HaveOccurred())
	return uuid.String()
}

func toJSON(input interface{}) io.Reader {
	b, err := json.Marshal(input)
	Expect(err).NotTo(HaveOccurred())
	return bytes.NewReader(b)
}

func fromJSON(input io.ReadCloser, output interface{}) {
	b, err := ioutil.ReadAll(input)
	Expect(err).NotTo(HaveOccurred())

	err = json.Unmarshal(b, output)
	Expect(err).NotTo(HaveOccurred())
}
