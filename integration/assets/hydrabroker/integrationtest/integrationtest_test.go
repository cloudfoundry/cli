package integrationtest_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/app"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	uuid2 "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
)

var _ = Describe("Integration Test", func() {
	var (
		server      *httptest.Server
		client      *http.Client
		create      func(io.Reader) string
		httpRequest func(cfg config.BrokerConfiguration, method, url string, body io.Reader) *http.Response
	)

	BeforeEach(func() {
		server = httptest.NewServer(app.App())
		client = server.Client()
		create = creator(server, client)
		httpRequest = requestor(client)
	})

	AfterEach(func() {
		server.Close()
	})

	It("responds to an aliveness test", func() {
		response, err := client.Head(server.URL)
		Expect(err).NotTo(HaveOccurred())
		expectStatusCode(response, http.StatusNoContent)
	})

	It("allows a broker to be created", func() {
		create(toJSON(randomConfiguration()))
	})

	When("the create request is missing parameters", func() {
		It("fails", func() {
			response, err := client.Post(server.URL+"/config", "application/json", strings.NewReader("{}"))
			Expect(err).NotTo(HaveOccurred())
			expectStatusCode(response, http.StatusBadRequest)

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
			guid = create(toJSON(cfg))
		})

		It("lists the broker", func() {
			response, err := client.Get(server.URL + "/config")
			Expect(err).NotTo(HaveOccurred())
			expectStatusCode(response, http.StatusOK)

			var r []string
			fromJSON(response.Body, &r)
			Expect(r).To(HaveLen(1))
			Expect(r[0]).To(Equal(guid))
		})

		It("rejects requests without a password", func() {
			response, err := client.Get(server.URL + "/broker/" + guid + "/v2/catalog")
			Expect(err).NotTo(HaveOccurred())
			expectStatusCode(response, http.StatusUnauthorized)
		})

		It("rejects requests with the wrong username", func() {
			cfg.Username = "wrong"
			response := httpRequest(cfg, "GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
			expectStatusCode(response, http.StatusUnauthorized)
		})

		It("rejects requests with the wrong password", func() {
			cfg.Password = "wrong"
			response := httpRequest(cfg, "GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
			expectStatusCode(response, http.StatusUnauthorized)
		})

		It("responds to the catalog endpoint", func() {
			response := httpRequest(cfg, "GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
			expectStatusCode(response, http.StatusOK)

			var catalog apiresponses.CatalogResponse
			fromJSON(response.Body, &catalog)
			Expect(catalog.Services).To(HaveLen(1))
			Expect(catalog.Services[0].ID).To(Equal(cfg.Services[0].ID))
			Expect(catalog.Services[0].Name).To(Equal(cfg.Services[0].Name))
			Expect(catalog.Services[0].Description).To(Equal(cfg.Services[0].Description))
			Expect(catalog.Services[0].Plans).To(HaveLen(2))
			Expect(catalog.Services[0].Plans[0].ID).To(Equal(cfg.Services[0].Plans[0].ID))
			Expect(catalog.Services[0].Plans[0].Name).To(Equal(cfg.Services[0].Plans[0].Name))
			Expect(catalog.Services[0].Plans[0].Description).To(Equal(cfg.Services[0].Plans[0].Description))
			Expect(catalog.Services[0].Plans[0].Free).To(PointTo(Equal(cfg.Services[0].Plans[0].Free)))
			Expect(catalog.Services[0].Plans[0].Bindable).To(PointTo(Equal(cfg.Services[0].Bindable)))
			Expect(catalog.Services[0].Plans[1].ID).To(Equal(cfg.Services[0].Plans[1].ID))
			Expect(catalog.Services[0].Plans[1].Name).To(Equal(cfg.Services[0].Plans[1].Name))
			Expect(catalog.Services[0].Plans[1].Description).To(Equal(cfg.Services[0].Plans[1].Description))
			Expect(catalog.Services[0].Plans[1].Free).To(PointTo(Equal(cfg.Services[0].Plans[1].Free)))
			Expect(catalog.Services[0].Plans[1].Bindable).To(PointTo(Equal(cfg.Services[0].Bindable)))
		})

		It("allows a service instance to be created", func() {
			instanceGUID := randomString()
			response := httpRequest(cfg, "PUT", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, nil)
			expectStatusCode(response, http.StatusCreated)

			var r map[string]interface{}
			fromJSON(response.Body, &r)
			Expect(r).To(Equal(map[string]interface{}{
				"dashboard_url": "http://example.com",
			}))
		})

		It("allows a service instance to be deleted", func() {
			instanceGUID := randomString()
			response := httpRequest(cfg, "DELETE", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, nil)
			expectStatusCode(response, http.StatusOK)

			var r map[string]interface{}
			fromJSON(response.Body, &r)
			Expect(r).To(BeEmpty())
		})

		It("allows a service instance to be updated", func() {
			instanceGUID := randomString()

			cfg.Services[0].Plans[0].MaintenanceInfo = &config.MaintenanceInfo{Version: "1.2.3"}
			response := httpRequest(cfg, "PATCH", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, toJSON(cfg))
			expectStatusCode(response, http.StatusOK)
		})

		It("allows a binding to be created", func() {
			instanceGUID := randomString()
			bindingGUID := randomString()
			response := httpRequest(cfg, "PUT", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID, nil)
			expectStatusCode(response, http.StatusCreated)

			var r map[string]interface{}
			fromJSON(response.Body, &r)
			Expect(r).To(BeEmpty())
		})

		It("allows a binding to be retrieved", func() {
			instanceGUID := randomString()
			bindingGUID := randomString()
			response := httpRequest(cfg, "GET", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID, nil)
			expectStatusCode(response, http.StatusOK)

			var r map[string]interface{}
			fromJSON(response.Body, &r)
			Expect(r).To(BeEmpty())
		})

		It("allows a binding to be deleted", func() {
			instanceGUID := randomString()
			bindingGUID := randomString()
			response := httpRequest(cfg, "DELETE", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID, nil)
			expectStatusCode(response, http.StatusOK)

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
				expectStatusCode(response, http.StatusNoContent)
			})

			By("no longer responding to the catalog endpoint", func() {
				response, err := client.Get(server.URL + "/broker/" + guid + "/v2/catalog")
				Expect(err).NotTo(HaveOccurred())
				expectStatusCode(response, http.StatusNotFound)
			})

			By("no longer listing the broker", func() {
				response, err := client.Get(server.URL + "/config")
				Expect(err).NotTo(HaveOccurred())
				expectStatusCode(response, http.StatusOK)

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
				expectStatusCode(response, http.StatusNoContent)
			})

			By("updating the catalog", func() {
				response := httpRequest(newCfg, "GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
				expectStatusCode(response, http.StatusOK)

				var catalog apiresponses.CatalogResponse
				fromJSON(response.Body, &catalog)
				Expect(catalog.Services[0].Name).To(Equal(newCfg.Services[0].Name))
			})
		})
	})

	Describe("configuring response codes", func() {
		var (
			guid string
			cfg  config.BrokerConfiguration
		)

		BeforeEach(func() {
			cfg = randomConfiguration()

			cfg.CatalogResponse = http.StatusInternalServerError
			cfg.ProvisionResponse = http.StatusBadGateway
			cfg.DeprovisionResponse = http.StatusTeapot
			cfg.BindResponse = http.StatusConflict
			cfg.UnbindResponse = http.StatusExpectationFailed
			cfg.GetBindingResponse = http.StatusGone

			guid = create(toJSON(cfg))
		})

		It("allows configuration of the catalog response code", func() {
			response := httpRequest(cfg, "GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
			expectStatusCode(response, http.StatusInternalServerError)
		})

		It("allows configuration of the provision response code", func() {
			instanceGUID := randomString()
			response := httpRequest(cfg, "PUT", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, nil)
			expectStatusCode(response, http.StatusBadGateway)
		})

		It("allows configuration of the deprovision response code", func() {
			instanceGUID := randomString()
			response := httpRequest(cfg, "DELETE", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID, nil)
			expectStatusCode(response, http.StatusTeapot)
		})

		It("allows configuration of the bind response code", func() {
			instanceGUID := randomString()
			bindingGUID := randomString()
			response := httpRequest(cfg, "PUT", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID, nil)
			expectStatusCode(response, http.StatusConflict)
		})

		It("allows configuration of the get binding response code", func() {
			instanceGUID := randomString()
			bindingGUID := randomString()
			response := httpRequest(cfg, "GET", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID, nil)
			expectStatusCode(response, http.StatusGone)
		})

		It("allows configuration of the unbind response code", func() {
			instanceGUID := randomString()
			bindingGUID := randomString()
			response := httpRequest(cfg, "DELETE", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID, nil)
			expectStatusCode(response, http.StatusExpectationFailed)
		})
	})

	Describe("configuring async responses", func() {
		const delay = 100 * time.Millisecond

		var (
			guid string
			cfg  config.BrokerConfiguration
		)
		BeforeEach(func() {
			cfg = randomConfiguration()
			cfg.AsyncResponseDelay = delay
			guid = create(toJSON(cfg))
		})

		getLastOperation := func(cfg config.BrokerConfiguration, op, urlPath string) string {
			response := httpRequest(
				cfg,
				"GET",
				urlPath+"?operation="+url.QueryEscape(op),
				nil,
			)
			expectStatusCode(response, http.StatusOK)

			var lastOperationResponse apiresponses.LastOperationResponse
			fromJSON(response.Body, &lastOperationResponse)

			return string(lastOperationResponse.State)
		}

		getInstanceLastOperation := func(cfg config.BrokerConfiguration, guid, instanceGUID, op string) string {
			return getLastOperation(cfg, op, server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/last_operation")
		}

		getBindingLastOperation := func(cfg config.BrokerConfiguration, guid, instanceGUID, bindingGUID, op string) string {
			return getLastOperation(cfg, op, server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID+"/last_operation")
		}

		It("does async provision", func() {
			var operation string
			instanceGUID := randomString()

			By("accepting the request", func() {
				response := httpRequest(cfg, "PUT", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"?accepts_incomplete=true", nil)
				expectStatusCode(response, http.StatusAccepted)

				var provisionResponse apiresponses.ProvisioningResponse
				fromJSON(response.Body, &provisionResponse)
				Expect(provisionResponse.DashboardURL).To(Equal("http://example.com"))

				operation = provisionResponse.OperationData
			})

			By("responding that the operation is still in progress", func() {
				Expect(getInstanceLastOperation(cfg, guid, instanceGUID, operation)).To(Equal("in progress"))
			})

			time.Sleep(delay)

			By("responding that the operation is complete", func() {
				Expect(getInstanceLastOperation(cfg, guid, instanceGUID, operation)).To(Equal("succeeded"))
			})
		})

		It("does async deprovision", func() {
			var operation string
			instanceGUID := randomString()

			By("accepting the request", func() {
				response := httpRequest(cfg, "DELETE", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"?accepts_incomplete=true", nil)
				expectStatusCode(response, http.StatusAccepted)

				var provisionResponse apiresponses.ProvisioningResponse
				fromJSON(response.Body, &provisionResponse)
				operation = provisionResponse.OperationData
			})

			By("responding that the operation is still in progress", func() {
				Expect(getInstanceLastOperation(cfg, guid, instanceGUID, operation)).To(Equal("in progress"))
			})

			time.Sleep(delay)

			By("responding that the operation is complete", func() {
				Expect(getInstanceLastOperation(cfg, guid, instanceGUID, operation)).To(Equal("succeeded"))
			})
		})

		It("does async bind", func() {
			var operation string
			instanceGUID := randomString()
			bindingGUID := randomString()

			By("accepting the request", func() {
				response := httpRequest(cfg, "PUT", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID+"?accepts_incomplete=true", nil)
				expectStatusCode(response, http.StatusAccepted)

				var provisionResponse apiresponses.ProvisioningResponse
				fromJSON(response.Body, &provisionResponse)
				operation = provisionResponse.OperationData
			})

			By("responding that the operation is still in progress", func() {
				Expect(getBindingLastOperation(cfg, guid, instanceGUID, bindingGUID, operation)).To(Equal("in progress"))
			})

			time.Sleep(delay)

			By("responding that the operation is complete", func() {
				Expect(getBindingLastOperation(cfg, guid, instanceGUID, bindingGUID, operation)).To(Equal("succeeded"))
			})
		})

		It("does async unbind", func() {
			var operation string
			instanceGUID := randomString()
			bindingGUID := randomString()

			By("accepting the request", func() {
				response := httpRequest(cfg, "DELETE", server.URL+"/broker/"+guid+"/v2/service_instances/"+instanceGUID+"/service_bindings/"+bindingGUID+"?accepts_incomplete=true", nil)
				expectStatusCode(response, http.StatusAccepted)

				var provisionResponse apiresponses.ProvisioningResponse
				fromJSON(response.Body, &provisionResponse)
				operation = provisionResponse.OperationData
			})

			By("responding that the operation is still in progress", func() {
				Expect(getBindingLastOperation(cfg, guid, instanceGUID, bindingGUID, operation)).To(Equal("in progress"))
			})

			time.Sleep(delay)

			By("responding that the operation is complete", func() {
				Expect(getBindingLastOperation(cfg, guid, instanceGUID, bindingGUID, operation)).To(Equal("succeeded"))
			})
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

				guid = create(toJSON(cfg))
			})

			By("showing it in the catalog", func() {
				response := httpRequest(cfg, "GET", server.URL+"/broker/"+guid+"/v2/catalog", nil)
				expectStatusCode(response, http.StatusOK)

				var catalog apiresponses.CatalogResponse
				fromJSON(response.Body, &catalog)
				Expect(catalog.Services[0].Name).To(Equal(cfg.Services[0].Name))
				Expect(catalog.Services[0].Plans[0].MaintenanceInfo.Version).To(Equal("1.2.3"))
				Expect(catalog.Services[0].Plans[0].MaintenanceInfo.Description).To(Equal("a description"))
			})
		})
	})
})

func expectStatusCode(response *http.Response, statusCode int) {
	ExpectWithOffset(1, response.StatusCode).To(Equal(statusCode), func() string {
		b, err := ioutil.ReadAll(response.Body)
		if err == nil {
			response.Body.Close()
			return "Body: " + string(b)
		}
		return "no error message found in body"
	})
}

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
						Free:        true,
					},
					{
						Name:        randomString(),
						ID:          randomString(),
						Description: randomString(),
						Free:        false,
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

	err = input.Close()
	Expect(err).NotTo(HaveOccurred())

	err = json.Unmarshal(b, output)
	Expect(err).NotTo(HaveOccurred(), string(b))
}

func requestor(client *http.Client) func(config.BrokerConfiguration, string, string, io.Reader) *http.Response {
	return func(cfg config.BrokerConfiguration, method, url string, body io.Reader) *http.Response {
		request, err := http.NewRequest(method, url, body)
		Expect(err).NotTo(HaveOccurred())
		request.SetBasicAuth(cfg.Username, cfg.Password)
		response, err := client.Do(request)
		Expect(err).NotTo(HaveOccurred())
		return response
	}
}

func creator(server *httptest.Server, client *http.Client) func(io.Reader) string {
	return func(body io.Reader) string {
		response, err := client.Post(server.URL+"/config", "application/json", body)
		Expect(err).NotTo(HaveOccurred())
		expectStatusCode(response, http.StatusCreated)

		var r config.NewBrokerResponse
		fromJSON(response.Body, &r)
		Expect(r.GUID).NotTo(BeEmpty())
		return r.GUID
	}
}
