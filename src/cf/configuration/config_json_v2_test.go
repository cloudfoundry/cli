package configuration_test

import (
	. "cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parsing V2 config files", func() {
	Describe("when the file is not valid", func() {
		It("returns an error", func() {

		})
	})

	Describe("when the file is valid", func() {
		var json = []byte(`
			{
				"ConfigVersion": 2,
				"Target": "api.example.com",
				"ApiVersion": "2",
				"AuthorizationEndpoint": "auth.example.com",
				"LoggregatorEndpoint": "logs.example.com",
				"AccessToken": "the-access-token",
				"RefreshToken": "the-refresh-token",
				"OrganizationFields": {
					"Name": "the-org"
				},
				"SpaceFields": {
					"Name": "the-space"
				}
			}`)

		It("returns a populated config object", func() {
			configData := NewData()
			err := JsonUnmarshalV2(json, configData)

			Expect(err).NotTo(HaveOccurred())
			Expect(configData).To(Equal(&Data{
				Target:                "api.example.com",
				ApiVersion:            "2",
				AuthorizationEndpoint: "auth.example.com",
				LoggregatorEndPoint:   "logs.example.com",
				AccessToken:           "the-access-token",
				RefreshToken:          "the-refresh-token",
				OrganizationFields:    models.OrganizationFields{Name: "the-org"},
				SpaceFields:           models.SpaceFields{Name: "the-space"},
			}))
		})
	})
})
