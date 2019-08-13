package fakeservicebroker

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
)

type FakeServiceBroker struct {
	name      string
	Services  []service
	domain    string
	behaviors behaviors
	reusable  bool
}

func New() *FakeServiceBroker {
	integrationNameGenerator := func() string {
		return helpers.PrefixedRandomName("INTEGRATION-PLAN")
	}

	firstPlanName := integrationNameGenerator()

	return &FakeServiceBroker{
		name:     generateReusableBrokerName(""),
		reusable: true,
		domain:   helpers.DefaultSharedDomain(),
		Services: []service{
			{
				Name:                 helpers.PrefixedRandomName("INTEGRATION-SERVICE"),
				ID:                   helpers.RandomName(),
				Description:          helpers.PrefixedRandomName("fake service description"),
				Bindable:             true,
				InstancesRetrievable: true,
				BindingsRetrievable:  true,
				Metadata: metadata{
					Shareable:        true,
					DocumentationURL: "http://documentation.url",
				},
				Requires: []string{},
				Plans: []plan{
					{
						Name:        firstPlanName,
						ID:          helpers.RandomName(),
						Description: helpers.PrefixedRandomName("fake plan description"),
						MaintenanceInfo: &maintenanceInfo{
							Version:     "2.0.0",
							Description: "OS image update.\nExpect downtime.",
						},
					},
					{
						Name:        helpers.GenerateHigherName(integrationNameGenerator, firstPlanName),
						ID:          helpers.RandomName(),
						Description: helpers.PrefixedRandomName("fake plan description"),
						MaintenanceInfo: &maintenanceInfo{
							Version:     "2.0.0",
							Description: "OS image update.\nExpect downtime.",
						},
					},
				},
			},
		},
		behaviors: behaviors{
			Provision: map[string]responseMock{
				"default": syncResponse().
					withBody(map[string]interface{}{
						"dashboard_url": "http://example.com",
					}),
			},
			Fetch: map[string]map[string]responseMock{
				"default": {
					"in_progress": syncResponse().
						withBody(map[string]interface{}{
							"state":       "in progress",
							"description": "not 100 percent done",
						}),
					"finished": syncResponse().
						withBody(map[string]interface{}{
							"state":       "succeeded",
							"description": "100 percent done",
						}),
				},
			},
			Update: map[string]responseMock{
				"default": syncResponse(),
			},
			Deprovision: map[string]responseMock{
				"default": syncResponse(),
			},
			Bind: map[string]responseMock{
				"default": syncResponse().
					withStatus(http.StatusCreated).
					withBody(map[string]interface{}{
						"credentials": map[string]interface{}{
							"uri":      "fake-service://fake-user:fake-password@fake-host:3306/fake-dbname",
							"username": "fake-user",
							"password": "fake-password",
							"host":     "fake-host",
							"port":     3306,
							"database": "fake-dbname",
						},
					}),
			},
			FetchServiceBinding: map[string]responseMock{
				"default": syncResponse().
					withBody(map[string]interface{}{
						"credentials": map[string]interface{}{
							"uri":      "fake-service://fake-user:fake-password@fake-host:3306/fake-dbname",
							"username": "fake-user",
							"password": "fake-password",
							"host":     "fake-host",
							"port":     3306,
							"database": "fake-dbname",
						},
					}),
			},
			Unbind: map[string]responseMock{
				"default": syncResponse(),
			},
		},
	}
}

// NewAlternate returns a reusable broker with another name. Can be used in conjunction with New() if you need two brokers in the same test.
func NewAlternate() *FakeServiceBroker {
	f := New()
	f.name = generateReusableBrokerName("other-")
	return f
}

func (f *FakeServiceBroker) WithName(name string) *FakeServiceBroker {
	f.name = name
	f.reusable = false
	return f
}

func (f *FakeServiceBroker) Async() *FakeServiceBroker {
	f.behaviors.Provision["default"] = asyncResponse()
	f.behaviors.Update["default"] = asyncResponse()
	f.behaviors.Deprovision["default"] = asyncResponse()
	f.behaviors.Bind["default"] = asyncResponse().asyncOnly()
	f.behaviors.Unbind["default"] = asyncResponse().asyncOnly()
	return f
}

func (f *FakeServiceBroker) Deploy() *FakeServiceBroker {
	f.pushAppIfNecessary()
	f.configure()
	return f
}

func (f *FakeServiceBroker) Register() *FakeServiceBroker {
	f.Deploy()
	f.register()
	return f
}

func (f *FakeServiceBroker) Update() *FakeServiceBroker {
	f.configure()
	f.update()
	return f
}

func (f *FakeServiceBroker) Destroy() {
	f.deregister()
	f.deleteApp()
}

func (f *FakeServiceBroker) URL(paths ...string) string {
	return fmt.Sprintf("http://%s.%s%s", f.name, f.domain, strings.Join(paths, ""))
}

func (f *FakeServiceBroker) ServiceName() string {
	return f.Services[0].Name
}

func (f *FakeServiceBroker) ServiceDescription() string {
	return f.Services[0].Description
}

func (f *FakeServiceBroker) ServicePlanDescription() string {
	return f.Services[0].Plans[0].Description
}

func (f *FakeServiceBroker) ServicePlanName() string {
	return f.Services[0].Plans[0].Name
}

func (f *FakeServiceBroker) Name() string {
	return f.name
}

func Setup() {
	helpers.WithRandomHomeDir(func() {
		helpers.SetAPI()
		helpers.LoginCF()
		helpers.CreateOrgAndSpaceUnlessExists(itsOrg, itsSpace)
	})
}

func Cleanup() {
	helpers.WithRandomHomeDir(func() {
		helpers.SetAPI()
		helpers.LoginCF()
		helpers.CreateOrgAndSpaceUnlessExists(itsOrg, itsSpace)
		helpers.TargetOrgAndSpace(itsOrg, itsSpace)

		broker := New()
		otherBroker := NewAlternate()

		if os.Getenv("KEEP_FAKE_SERVICE_BROKERS") != "true" {
			broker.stopReusing()
			otherBroker.stopReusing()
		}

		broker.cleanup()
		otherBroker.cleanup()
	})
}
