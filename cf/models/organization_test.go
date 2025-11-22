package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization", func() {
	Describe("OrganizationFields", func() {
		It("stores basic organization information", func() {
			org := models.OrganizationFields{
				Guid: "org-guid",
				Name: "my-org",
				QuotaDefinition: models.QuotaFields{
					Guid: "quota-guid",
					Name: "default-quota",
				},
			}

			Expect(org.Guid).To(Equal("org-guid"))
			Expect(org.Name).To(Equal("my-org"))
			Expect(org.QuotaDefinition.Guid).To(Equal("quota-guid"))
		})
	})

	Describe("Organization", func() {
		It("embeds OrganizationFields", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Guid: "org-guid",
					Name: "my-org",
				},
			}

			Expect(org.Guid).To(Equal("org-guid"))
			Expect(org.Name).To(Equal("my-org"))
		})

		It("has spaces", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Guid: "org-guid",
					Name: "my-org",
				},
				Spaces: []models.SpaceFields{
					{Guid: "space-1-guid", Name: "space-1"},
					{Guid: "space-2-guid", Name: "space-2"},
				},
			}

			Expect(len(org.Spaces)).To(Equal(2))
			Expect(org.Spaces[0].Name).To(Equal("space-1"))
			Expect(org.Spaces[1].Name).To(Equal("space-2"))
		})

		It("has domains", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Guid: "org-guid",
					Name: "my-org",
				},
				Domains: []models.DomainFields{
					{Guid: "domain-1-guid", Name: "example.com"},
					{Guid: "domain-2-guid", Name: "test.com"},
				},
			}

			Expect(len(org.Domains)).To(Equal(2))
			Expect(org.Domains[0].Name).To(Equal("example.com"))
			Expect(org.Domains[1].Name).To(Equal("test.com"))
		})

		It("has space quotas", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Guid: "org-guid",
					Name: "my-org",
				},
				SpaceQuotas: []models.SpaceQuota{
					{Guid: "quota-1-guid", Name: "space-quota-1"},
					{Guid: "quota-2-guid", Name: "space-quota-2"},
				},
			}

			Expect(len(org.SpaceQuotas)).To(Equal(2))
			Expect(org.SpaceQuotas[0].Name).To(Equal("space-quota-1"))
		})

		It("can have empty collections", func() {
			org := models.Organization{
				OrganizationFields: models.OrganizationFields{
					Guid: "org-guid",
					Name: "my-org",
				},
				Spaces:      []models.SpaceFields{},
				Domains:     []models.DomainFields{},
				SpaceQuotas: []models.SpaceQuota{},
			}

			Expect(len(org.Spaces)).To(Equal(0))
			Expect(len(org.Domains)).To(Equal(0))
			Expect(len(org.SpaceQuotas)).To(Equal(0))
		})
	})
})
