package core_config_test

import (
	"regexp"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V3 Config files", func() {
	var exampleV3JSON = `
	{
		"ConfigVersion": 3,
		"Target": "api.example.com",
		"ApiVersion": "3",
		"AuthorizationEndpoint": "auth.example.com",
		"LoggregatorEndPoint": "loggregator.example.com",
		"DopplerEndPoint": "doppler.example.com",
		"UaaEndpoint": "uaa.example.com",
		"RoutingApiEndpoint": "routing-api.example.com",
		"AccessToken": "the-access-token",
		"SSHOAuthClient": "ssh-oauth-client-id",
		"RefreshToken": "the-refresh-token",
		"OrganizationFields": {
			"Guid": "the-org-guid",
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
			"Guid": "the-space-guid",
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
			"Url": "http://repo.com"
		}
		],
		"MinCliVersion": "6.0.0",
		"MinRecommendedCliVersion": "6.9.0"
	}`

	// V2 by virtue of ConfigVersion only
	var exampleV2JSON = `
	{
		"ConfigVersion": 2,
		"Target": "api.example.com",
		"ApiVersion": "3",
		"AuthorizationEndpoint": "auth.example.com",
		"LoggregatorEndPoint": "loggregator.example.com",
		"DopplerEndPoint": "doppler.example.com",
		"UaaEndpoint": "uaa.example.com",
		"RoutingApiEndpoint": "routing-api.example.com",
		"AccessToken": "the-access-token",
		"SSHOAuthClient": "ssh-oauth-client-id",
		"RefreshToken": "the-refresh-token",
		"OrganizationFields": {
			"Guid": "the-org-guid",
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
			"Guid": "the-space-guid",
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
			"Url": "http://repo.com"
		}
		],
		"MinCliVersion": "6.0.0",
		"MinRecommendedCliVersion": "6.9.0"
	}`

	Describe("JsonMarshalV3", func() {
		It("creates a JSON string from the config object", func() {
			data := &core_config.Data{
				Target:                   "api.example.com",
				ApiVersion:               "3",
				AuthorizationEndpoint:    "auth.example.com",
				LoggregatorEndPoint:      "loggregator.example.com",
				RoutingApiEndpoint:       "routing-api.example.com",
				DopplerEndPoint:          "doppler.example.com",
				UaaEndpoint:              "uaa.example.com",
				AccessToken:              "the-access-token",
				RefreshToken:             "the-refresh-token",
				SSHOAuthClient:           "ssh-oauth-client-id",
				MinCliVersion:            "6.0.0",
				MinRecommendedCliVersion: "6.9.0",
				OrganizationFields: models.OrganizationFields{
					Guid: "the-org-guid",
					Name: "the-org",
				},
				SpaceFields: models.SpaceFields{
					Guid: "the-space-guid",
					Name: "the-space",
				},
				SSLDisabled:  true,
				Trace:        "path/to/some/file",
				AsyncTimeout: 1000,
				ColorEnabled: "true",
				Locale:       "fr_FR",
				PluginRepos: []models.PluginRepo{
					models.PluginRepo{
						Name: "repo1",
						Url:  "http://repo.com",
					},
				},
			}

			jsonData, err := data.JsonMarshalV3()
			Expect(err).NotTo(HaveOccurred())

			re := regexp.MustCompile(`\s+`)
			actual := re.ReplaceAll(jsonData, []byte{})
			expected := re.ReplaceAll([]byte(exampleV3JSON), []byte{})
			Expect(actual).To(Equal(expected))
		})
	})

	Describe("JsonUnmarshalV3", func() {
		It("returns an error when the JSON is invalid", func() {
			configData := core_config.NewData()
			err := configData.JsonUnmarshalV3([]byte(`{ "not_valid": ### }`))
			Expect(err).To(HaveOccurred())
		})

		It("creates a config object from valid V3 JSON", func() {
			expectedData := &core_config.Data{
				ConfigVersion:            3,
				Target:                   "api.example.com",
				ApiVersion:               "3",
				AuthorizationEndpoint:    "auth.example.com",
				LoggregatorEndPoint:      "loggregator.example.com",
				RoutingApiEndpoint:       "routing-api.example.com",
				DopplerEndPoint:          "doppler.example.com",
				UaaEndpoint:              "uaa.example.com",
				AccessToken:              "the-access-token",
				RefreshToken:             "the-refresh-token",
				SSHOAuthClient:           "ssh-oauth-client-id",
				MinCliVersion:            "6.0.0",
				MinRecommendedCliVersion: "6.9.0",
				OrganizationFields: models.OrganizationFields{
					Guid: "the-org-guid",
					Name: "the-org",
				},
				SpaceFields: models.SpaceFields{
					Guid: "the-space-guid",
					Name: "the-space",
				},
				SSLDisabled:  true,
				Trace:        "path/to/some/file",
				AsyncTimeout: 1000,
				ColorEnabled: "true",
				Locale:       "fr_FR",
				PluginRepos: []models.PluginRepo{
					models.PluginRepo{
						Name: "repo1",
						Url:  "http://repo.com",
					},
				},
			}

			actualData := core_config.NewData()
			err := actualData.JsonUnmarshalV3([]byte(exampleV3JSON))
			Expect(err).NotTo(HaveOccurred())

			Expect(actualData).To(Equal(expectedData))
		})

		It("returns an empty Data object for non-V3 JSON", func() {
			actualData := core_config.NewData()
			err := actualData.JsonUnmarshalV3([]byte(exampleV2JSON))
			Expect(err).NotTo(HaveOccurred())

			Expect(*actualData).To(Equal(core_config.Data{}))
		})
	})
})
