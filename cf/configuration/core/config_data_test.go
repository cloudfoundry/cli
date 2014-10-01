package configuration_test

import (
	"regexp"

	. "github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var exampleJSON = `
{
	"ConfigVersion": 3,
	"Target": "api.example.com",
	"ApiVersion": "3",
	"AuthorizationEndpoint": "auth.example.com",
	"LoggregatorEndPoint": "logs.example.com",
	"UaaEndpoint": "uaa.example.com",
	"AccessToken": "the-access-token",
	"RefreshToken": "the-refresh-token",
	"OrganizationFields": {
		"Guid": "the-org-guid",
		"Name": "the-org",
		"QuotaDefinition": {
			"name":"",
			"memory_limit":0,
			"total_routes":0,
			"total_services":0,
			"non_basic_services_allowed": false
		}
	},
	"SpaceFields": {
		"Guid": "the-space-guid",
		"Name": "the-space"
	},
	"SSLDisabled": true,
	"AsyncTimeout": 1000,
	"Trace": "path/to/some/file",
	"ColorEnabled": "true",
	"Locale": "fr_FR"
}`

var exampleData = &Data{
	Target:                "api.example.com",
	ApiVersion:            "3",
	AuthorizationEndpoint: "auth.example.com",
	LoggregatorEndPoint:   "logs.example.com",
	UaaEndpoint:           "uaa.example.com",
	AccessToken:           "the-access-token",
	RefreshToken:          "the-refresh-token",
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
}

var _ = Describe("V3 Config files", func() {
	Describe("serialization", func() {
		It("creates a JSON string from the config object", func() {
			jsonData, err := exampleData.JsonMarshalV3()
			Expect(err).NotTo(HaveOccurred())

			jData := stripWhitespace(string(jsonData))
			strippedJson := stripWhitespace(exampleJSON)
			Expect(jData).To(ContainSubstring(strippedJson))
		})
	})

	Describe("parsing", func() {
		It("returns an error when the JSON is invalid", func() {
			configData := NewData()
			err := configData.JsonUnmarshalV3([]byte(`{ "not_valid": ### }`))

			Expect(err).To(HaveOccurred())
		})

		It("creates a config object from valid JSON", func() {
			configData := NewData()
			err := configData.JsonUnmarshalV3([]byte(exampleJSON))

			Expect(err).NotTo(HaveOccurred())
			Expect(configData).To(Equal(exampleData))
		})
	})
})

var whiteSpaceRegex = regexp.MustCompile(`\s+`)

func stripWhitespace(input string) string {
	return whiteSpaceRegex.ReplaceAllString(input, "")
}
