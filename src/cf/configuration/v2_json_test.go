package configuration_test

import (
	"cf"
	. "cf/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	Describe("Parsing V2 config files", func() {
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
				},
				"ApplicationStartTimeout": 5
			}`)

			It("returns a populated config object", func() {
				config, err := ConfigFromJsonV2(json)
				Expect(err).NotTo(HaveOccurred())

				Expect(*config.GetOldConfig()).To(Equal(Configuration{
					Target:                  "api.example.com",
					ApiVersion:              "2",
					AuthorizationEndpoint:   "auth.example.com",
					LoggregatorEndPoint:     "logs.example.com",
					AccessToken:             "the-access-token",
					RefreshToken:            "the-refresh-token",
					OrganizationFields:      cf.OrganizationFields{BasicFields: cf.BasicFields{Name: "the-org"}},
					SpaceFields:             cf.SpaceFields{BasicFields: cf.BasicFields{Name: "the-space"}},
					ApplicationStartTimeout: 5,
				}))
			})
		})
	})
}
