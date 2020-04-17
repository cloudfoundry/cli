package servicebrokerstub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/gomega"
)

func newDefaultConfig() *ServiceBrokerStub {
	return &ServiceBrokerStub{
		URL:      "broker-route-not-created-yet",
		Name:     helpers.NewServiceBrokerName(),
		Username: "username",
		Password: "password",
		Services: []config.Service{newDefaultServiceOffering()},
	}
}

func newDefaultServiceOffering() config.Service {
	return config.Service{
		Name:        helpers.NewServiceOfferingName(),
		ID:          helpers.RandomName(),
		Description: helpers.PrefixedRandomName("SERVICE-OFFERING-DESCRIPTION"),
		Shareable:   true,
		Bindable:    true,
		Plans:       []config.Plan{newDefaultPlan()},
	}
}

func newDefaultPlan() config.Plan {
	return config.Plan{
		Name:        helpers.NewPlanName(),
		ID:          helpers.RandomName(),
		Description: helpers.PrefixedRandomName("PLAN-DESCRIPTION"),
		Free:        true,
	}
}

func (s *ServiceBrokerStub) requestNewBrokerRoute() {
	req, err := http.NewRequest("POST", appURL("/config"), s.marshal())
	Expect(err).ToNot(HaveOccurred())

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	Expect(err).ToNot(HaveOccurred())

	responseBody, err := ioutil.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())
	Expect(resp.StatusCode).To(
		Equal(http.StatusCreated),
		fmt.Sprintf("Expected POST /config to succeed. Response body: '%s'", string(responseBody)),
	)
	defer resp.Body.Close()

	var response config.NewBrokerResponse
	err = json.Unmarshal(responseBody, &response)
	Expect(err).ToNot(HaveOccurred())

	s.URL = appURL("/broker/", response.GUID)
	s.GUID = response.GUID
}

func (s *ServiceBrokerStub) configure() {
	req, err := http.NewRequest("PUT", appURL("/config/"+s.GUID), s.marshal())
	Expect(err).ToNot(HaveOccurred())

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	Expect(err).ToNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
}

func (s *ServiceBrokerStub) forget() {
	req, err := http.NewRequest("DELETE", appURL("/config/", s.GUID), nil)
	Expect(err).ToNot(HaveOccurred())
	resp, err := http.DefaultClient.Do(req)
	resp.Body.Close()
	Expect(err).ToNot(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
}

func (s *ServiceBrokerStub) marshal() io.Reader {
	body, err := json.Marshal(config.BrokerConfiguration{
		Services:            s.Services,
		Username:            s.Username,
		Password:            s.Password,
		CatalogResponse:     s.catalogResponse,
		ProvisionResponse:   s.provisionResponse,
		DeprovisionResponse: s.deprovisionResponse,
		AsyncResponseDelay:  s.asyncResponseDelay,
	})
	Expect(err).ToNot(HaveOccurred())
	return bytes.NewReader(body)
}
