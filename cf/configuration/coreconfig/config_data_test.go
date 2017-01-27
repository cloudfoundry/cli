package coreconfig_test

import (
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V3 Config files", func() {
	var exampleV3JSON = `
	{
		"ConfigVersion": 3,
		"Target": "api.example.com",
		"APIVersion": "3",
		"AuthorizationEndpoint": "auth.example.com",
		"DopplerEndPoint": "doppler.example.com",
		"UaaEndpoint": "uaa.example.com",
		"RoutingAPIEndpoint": "routing-api.example.com",
		"AccessToken": "the-access-token",
		"UAAOAuthClient": "cf-oauth-client-id",
		"UAAOAuthClientSecret": "cf-oauth-client-secret",
		"SSHOAuthClient": "ssh-oauth-client-id",
		"RefreshToken": "the-refresh-token",
		"OrganizationFields": {
			"GUID": "the-org-guid",
			"Name": "the-org",
			"QuotaDefinition": {
				"name":"",
				"memory_limit":0,
				"instance_memory_limit":0,
				"total_routes":0,
				"total_services":0,
				"non_basic_services_allowed": false,
				"app_instance_limit":0
			}
		},
		"SpaceFields": {
			"GUID": "the-space-guid",
			"Name": "the-space",
			"AllowSSH": false
		},
		"SSLDisabled": true,
		"AsyncTimeout": 1000,
		"Trace": "path/to/some/file",
		"ColorEnabled": "true",
		"Locale": "fr_FR",
		"PluginRepos": [
		{
			"Name": "repo1",
			"URL": "http://repo.com"
		}
		],
		"MinCLIVersion": "6.0.0",
		"MinRecommendedCLIVersion": "6.9.0"
	}`

	// V2 by virtue of ConfigVersion only
	var exampleV2JSON = `
	{
		"ConfigVersion": 2,
		"Target": "api.example.com",
		"APIVersion": "3",
		"AuthorizationEndpoint": "auth.example.com",
		"LoggregatorEndPoint": "loggregator.example.com",
		"DopplerEndPoint": "doppler.example.com",
		"UaaEndpoint": "uaa.example.com",
		"RoutingAPIEndpoint": "routing-api.example.com",
		"AccessToken": "the-access-token",
		"UAAOAuthClient": "cf-oauth-client-id",
		"UAAOAuthClientSecret": "cf-oauth-client-secret",
		"SSHOAuthClient": "ssh-oauth-client-id",
		"RefreshToken": "the-refresh-token",
		"OrganizationFields": {
			"GUID": "the-org-guid",
			"Name": "the-org",
			"QuotaDefinition": {
				"name":"",
				"memory_limit":0,
				"instance_memory_limit":0,
				"total_routes":0,
				"total_services":0,
				"non_basic_services_allowed": false
			}
		},
		"SpaceFields": {
			"GUID": "the-space-guid",
			"Name": "the-space",
			"AllowSSH": false
		},
		"SSLDisabled": true,
		"AsyncTimeout": 1000,
		"Trace": "path/to/some/file",
		"ColorEnabled": "true",
		"Locale": "fr_FR",
		"PluginRepos": [
		{
			"Name": "repo1",
			"URL": "http://repo.com"
		}
		],
		"MinCLIVersion": "6.0.0",
		"MinRecommendedCLIVersion": "6.9.0"
	}`

	Describe("NewData", func() {
		It("sets default values for UAAOAuthClient and CFOAUthCLientSecret", func() {
			data := coreconfig.NewData()
			Expect(data.UAAOAuthClient).To(Equal("cf"))
			Expect(data.UAAOAuthClientSecret).To(Equal(""))
		})
	})

	Describe("JSONMarshalV3", func() {
		It("creates a JSON string from the config object", func() {
			data := &coreconfig.Data{
				Target:                   "api.example.com",
				APIVersion:               "3",
				AuthorizationEndpoint:    "auth.example.com",
				RoutingAPIEndpoint:       "routing-api.example.com",
				DopplerEndPoint:          "doppler.example.com",
				UaaEndpoint:              "uaa.example.com",
				AccessToken:              "the-access-token",
				RefreshToken:             "the-refresh-token",
				UAAOAuthClient:           "cf-oauth-client-id",
				UAAOAuthClientSecret:     "cf-oauth-client-secret",
				SSHOAuthClient:           "ssh-oauth-client-id",
				MinCLIVersion:            "6.0.0",
				MinRecommendedCLIVersion: "6.9.0",
				OrganizationFields: models.OrganizationFields{
					GUID: "the-org-guid",
					Name: "the-org",
				},
				SpaceFields: models.SpaceFields{
					GUID: "the-space-guid",
					Name: "the-space",
				},
				SSLDisabled:  true,
				Trace:        "path/to/some/file",
				AsyncTimeout: 1000,
				ColorEnabled: "true",
				Locale:       "fr_FR",
				PluginRepos: []models.PluginRepo{
					{
						Name: "repo1",
						URL:  "http://repo.com",
					},
				},
			}

			jsonData, err := data.JSONMarshalV3()
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonData).To(MatchJSON(exampleV3JSON))
		})
	})

	Describe("JSONUnmarshalV3", func() {
		It("returns an error when the JSON is invalid", func() {
			configData := coreconfig.NewData()
			err := configData.JSONUnmarshalV3([]byte(`{ "not_valid": ### }`))
			Expect(err).To(HaveOccurred())
		})

		It("creates a config object from valid V3 JSON", func() {
			expectedData := &coreconfig.Data{
				ConfigVersion:            3,
				Target:                   "api.example.com",
				APIVersion:               "3",
				AuthorizationEndpoint:    "auth.example.com",
				RoutingAPIEndpoint:       "routing-api.example.com",
				DopplerEndPoint:          "doppler.example.com",
				UaaEndpoint:              "uaa.example.com",
				AccessToken:              "the-access-token",
				RefreshToken:             "the-refresh-token",
				UAAOAuthClient:           "cf-oauth-client-id",
				UAAOAuthClientSecret:     "cf-oauth-client-secret",
				SSHOAuthClient:           "ssh-oauth-client-id",
				MinCLIVersion:            "6.0.0",
				MinRecommendedCLIVersion: "6.9.0",
				OrganizationFields: models.OrganizationFields{
					GUID: "the-org-guid",
					Name: "the-org",
				},
				SpaceFields: models.SpaceFields{
					GUID: "the-space-guid",
					Name: "the-space",
				},
				SSLDisabled:  true,
				Trace:        "path/to/some/file",
				AsyncTimeout: 1000,
				ColorEnabled: "true",
				Locale:       "fr_FR",
				PluginRepos: []models.PluginRepo{
					{
						Name: "repo1",
						URL:  "http://repo.com",
					},
				},
			}

			actualData := coreconfig.NewData()
			err := actualData.JSONUnmarshalV3([]byte(exampleV3JSON))
			Expect(err).NotTo(HaveOccurred())

			Expect(actualData).To(Equal(expectedData))
		})

		It("returns an empty Data object for non-V3 JSON", func() {
			actualData := coreconfig.NewData()
			err := actualData.JSONUnmarshalV3([]byte(exampleV2JSON))
			Expect(err).NotTo(HaveOccurred())

			Expect(*actualData).To(Equal(coreconfig.Data{}))
		})
	})
})
