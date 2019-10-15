package fakeservicebroker

import (
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
)

func defaultConfig() *FakeServiceBroker {
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
