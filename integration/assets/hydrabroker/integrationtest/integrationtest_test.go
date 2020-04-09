package integrationtest_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	uuid2 "github.com/nu7hatch/gouuid"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/app"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"net/http/httptest"
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
	})
})

func randomConfiguration() config.BrokerConfiguration {
	return config.BrokerConfiguration{
		Services: []config.Service{
			{
				Name: randomString(),
				Plans: []config.Plan{
					{
						Name: randomString(),
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
