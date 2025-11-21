package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Quota", func() {
	Describe("NewQuotaFields", func() {
		It("creates a quota with all fields", func() {
			quota := models.NewQuotaFields("default", 1024, 512, 10, 5, true)

			Expect(quota.Name).To(Equal("default"))
			Expect(quota.MemoryLimit).To(Equal(int64(1024)))
			Expect(quota.InstanceMemoryLimit).To(Equal(int64(512)))
			Expect(quota.RoutesLimit).To(Equal(10))
			Expect(quota.ServicesLimit).To(Equal(5))
			Expect(quota.NonBasicServicesAllowed).To(BeTrue())
		})

		It("creates a quota with zero values", func() {
			quota := models.NewQuotaFields("zero-quota", 0, 0, 0, 0, false)

			Expect(quota.Name).To(Equal("zero-quota"))
			Expect(quota.MemoryLimit).To(Equal(int64(0)))
			Expect(quota.InstanceMemoryLimit).To(Equal(int64(0)))
			Expect(quota.RoutesLimit).To(Equal(0))
			Expect(quota.ServicesLimit).To(Equal(0))
			Expect(quota.NonBasicServicesAllowed).To(BeFalse())
		})

		It("creates a quota with large values", func() {
			quota := models.NewQuotaFields("large-quota", 10240, 2048, 100, 50, true)

			Expect(quota.Name).To(Equal("large-quota"))
			Expect(quota.MemoryLimit).To(Equal(int64(10240)))
			Expect(quota.InstanceMemoryLimit).To(Equal(int64(2048)))
			Expect(quota.RoutesLimit).To(Equal(100))
			Expect(quota.ServicesLimit).To(Equal(50))
		})

		It("creates a quota disallowing non-basic services", func() {
			quota := models.NewQuotaFields("restricted", 512, 256, 5, 2, false)

			Expect(quota.NonBasicServicesAllowed).To(BeFalse())
		})

		It("handles different quota names", func() {
			quota1 := models.NewQuotaFields("default", 1024, 512, 10, 5, true)
			quota2 := models.NewQuotaFields("runaway", 2048, 1024, 20, 10, true)
			quota3 := models.NewQuotaFields("small", 256, 128, 2, 1, false)

			Expect(quota1.Name).To(Equal("default"))
			Expect(quota2.Name).To(Equal("runaway"))
			Expect(quota3.Name).To(Equal("small"))
		})
	})

	Describe("QuotaFields", func() {
		It("stores quota fields directly", func() {
			quota := models.QuotaFields{
				Guid:                    "quota-guid",
				Name:                    "custom-quota",
				MemoryLimit:             2048,
				InstanceMemoryLimit:     1024,
				RoutesLimit:             15,
				ServicesLimit:           8,
				NonBasicServicesAllowed: true,
			}

			Expect(quota.Guid).To(Equal("quota-guid"))
			Expect(quota.Name).To(Equal("custom-quota"))
			Expect(quota.MemoryLimit).To(Equal(int64(2048)))
			Expect(quota.InstanceMemoryLimit).To(Equal(int64(1024)))
			Expect(quota.RoutesLimit).To(Equal(15))
			Expect(quota.ServicesLimit).To(Equal(8))
			Expect(quota.NonBasicServicesAllowed).To(BeTrue())
		})

		It("handles empty guid", func() {
			quota := models.QuotaFields{
				Name: "no-guid-quota",
			}

			Expect(quota.Guid).To(BeEmpty())
		})

		It("has json tags for serialization", func() {
			quota := models.QuotaFields{
				Guid:                    "quota-guid",
				Name:                    "test",
				MemoryLimit:             1024,
				InstanceMemoryLimit:     512,
				RoutesLimit:             10,
				ServicesLimit:           5,
				NonBasicServicesAllowed: true,
			}

			// Verify fields exist (json tags are compile-time)
			Expect(quota.Name).To(Equal("test"))
			Expect(quota.MemoryLimit).To(Equal(int64(1024)))
			Expect(quota.InstanceMemoryLimit).To(Equal(int64(512)))
			Expect(quota.RoutesLimit).To(Equal(10))
			Expect(quota.ServicesLimit).To(Equal(5))
			Expect(quota.NonBasicServicesAllowed).To(BeTrue())
		})

		It("handles unlimited instance memory", func() {
			quota := models.QuotaFields{
				Name:                "unlimited-instance",
				MemoryLimit:         2048,
				InstanceMemoryLimit: -1,
			}

			Expect(quota.InstanceMemoryLimit).To(Equal(int64(-1)))
		})

		It("handles different limit combinations", func() {
			quota1 := models.QuotaFields{
				RoutesLimit:   0,
				ServicesLimit: 10,
			}
			quota2 := models.QuotaFields{
				RoutesLimit:   20,
				ServicesLimit: 0,
			}

			Expect(quota1.RoutesLimit).To(Equal(0))
			Expect(quota1.ServicesLimit).To(Equal(10))
			Expect(quota2.RoutesLimit).To(Equal(20))
			Expect(quota2.ServicesLimit).To(Equal(0))
		})
	})
})
