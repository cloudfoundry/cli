package resources_test

import (
	. "code.cloudfoundry.org/cli/cf/api/resources"

	"code.cloudfoundry.org/cli/cf/models"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Quotas", func() {
	Describe("ToFields", func() {
		var resource QuotaResource

		BeforeEach(func() {
			resource = QuotaResource{
				Resource: Resource{
					Metadata: Metadata{
						GUID: "my-guid",
						URL:  "url.com",
					},
				},
				Entity: models.QuotaResponse{
					GUID:                    "my-guid",
					Name:                    "my-name",
					MemoryLimit:             1024,
					InstanceMemoryLimit:     5,
					RoutesLimit:             10,
					ServicesLimit:           5,
					NonBasicServicesAllowed: true,
					AppInstanceLimit:        "10",
				},
			}
		})

		Describe("ReservedRoutePorts", func() {
			Context("when it is provided by the API", func() {
				BeforeEach(func() {
					resource.Entity.ReservedRoutePorts = "5"
				})

				It("returns back the value", func() {
					fields := resource.ToFields()
					Expect(fields.ReservedRoutePorts).To(Equal(json.Number("5")))
				})
			})

			Context("when it is *not* provided by the API", func() {
				It("should be empty", func() {
					fields := resource.ToFields()
					Expect(fields.ReservedRoutePorts).To(BeEmpty())
				})
			})
		})
	})
})
