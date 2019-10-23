package fakeservicebroker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/gomega"
)

func (f *FakeServiceBroker) configure() {
	Eventually(f.alreadyPushedApp).Should(
		Equal(true),
		fmt.Sprintf("Expected app to be pushed and eventually available on %s", f.URL()),
	)

	req, err := http.NewRequest("POST", f.URL("/config"), f.configJSON())
	Expect(err).ToNot(HaveOccurred())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	Expect(err).ToNot(HaveOccurred())
	body, err := ioutil.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())
	Expect(resp.StatusCode).To(
		Equal(http.StatusOK),
		fmt.Sprintf("Expected POST /config to succeed. Response body: '%s'", string(body)),
	)
	defer resp.Body.Close()
}

func (f *FakeServiceBroker) configJSON() io.Reader {
	f.behaviors.Catalog = syncResponse().
		withBody(map[string]interface{}{
			"services": f.Services,
		}).
		withStatus(f.catalogStatus)

	config := configuration{Behaviors: f.behaviors}

	data, err := json.Marshal(config)
	Expect(err).NotTo(HaveOccurred())
	return bytes.NewReader(data)
}

func asyncResponse() responseMock {
	return responseMock{
		SleepSeconds: 0,
		Status:       http.StatusAccepted,
		Body:         map[string]interface{}{},
	}
}

func syncResponse() responseMock {
	return responseMock{
		SleepSeconds: 0,
		Status:       http.StatusOK,
		Body:         map[string]interface{}{},
	}
}

func (r responseMock) asyncOnly() responseMock {
	var nillableTrue = new(bool)
	*nillableTrue = true
	r.AsyncOnly = nillableTrue
	return r
}

func (r responseMock) withBody(body map[string]interface{}) responseMock {
	r.Body = body
	return r
}

func (r responseMock) withStatus(status int) responseMock {
	r.Status = status
	return r
}

type configuration struct {
	Behaviors behaviors `json:"behaviors"`
}

type behaviors struct {
	Catalog             responseMock                       `json:"catalog"`
	Provision           map[string]responseMock            `json:"provision"`
	Bind                map[string]responseMock            `json:"bind"`
	Fetch               map[string]map[string]responseMock `json:"fetch"`
	Update              map[string]responseMock            `json:"update"`
	Deprovision         map[string]responseMock            `json:"deprovision"`
	Unbind              map[string]responseMock            `json:"unbind"`
	FetchServiceBinding map[string]responseMock            `json:"fetch_service_binding"`
}

type service struct {
	Name                 string   `json:"name"`
	ID                   string   `json:"id"`
	Description          string   `json:"description"`
	Bindable             bool     `json:"bindable"`
	InstancesRetrievable bool     `json:"instances_retrievable"`
	BindingsRetrievable  bool     `json:"bindings_retrievable"`
	Plans                []plan   `json:"plans"`
	Requires             []string `json:"requires"`
	Metadata             metadata `json:"metadata"`
}

func (s service) PlanNames() []string {
	var planNames []string
	for _, plan := range s.Plans {
		planNames = append(planNames, plan.Name)
	}
	return planNames
}

type plan struct {
	Name            string           `json:"name"`
	ID              string           `json:"id"`
	Description     string           `json:"description"`
	MaintenanceInfo *maintenanceInfo `json:"maintenance_info,omitempty"`
}

func (p *plan) SetMaintenanceInfo(version, description string) {
	p.MaintenanceInfo.Version = version
	p.MaintenanceInfo.Description = description
}

func (p *plan) RemoveMaintenanceInfo() {
	p.MaintenanceInfo = nil
}

type maintenanceInfo struct {
	Version     string `json:"version"`
	Description string `json:"description"`
}

type responseMock struct {
	SleepSeconds int                    `json:"sleep_seconds"`
	Status       int                    `json:"status"`
	Body         map[string]interface{} `json:"body"`
	AsyncOnly    *bool                  `json:"async_only,omitempty"`
}

type metadata struct {
	Shareable        bool   `json:"shareable"`
	DocumentationURL string `json:"documentationUrl"`
}
