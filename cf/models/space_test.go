package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space", func() {
	Describe("SpaceFields", func() {
		It("stores basic space information", func() {
			space := models.SpaceFields{
				Guid: "space-guid",
				Name: "my-space",
			}

			Expect(space.Guid).To(Equal("space-guid"))
			Expect(space.Name).To(Equal("my-space"))
		})

		It("handles empty values", func() {
			space := models.SpaceFields{}

			Expect(space.Guid).To(BeEmpty())
			Expect(space.Name).To(BeEmpty())
		})
	})

	Describe("Space", func() {
		It("embeds SpaceFields", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
			}

			Expect(space.Guid).To(Equal("space-guid"))
			Expect(space.Name).To(Equal("my-space"))
		})

		It("has an organization", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
				Organization: models.OrganizationFields{
					Guid: "org-guid",
					Name: "my-org",
				},
			}

			Expect(space.Organization.Guid).To(Equal("org-guid"))
			Expect(space.Organization.Name).To(Equal("my-org"))
		})

		It("has applications", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
				Applications: []models.ApplicationFields{
					{Guid: "app-1-guid", Name: "app-1"},
					{Guid: "app-2-guid", Name: "app-2"},
					{Guid: "app-3-guid", Name: "app-3"},
				},
			}

			Expect(len(space.Applications)).To(Equal(3))
			Expect(space.Applications[0].Name).To(Equal("app-1"))
			Expect(space.Applications[1].Name).To(Equal("app-2"))
			Expect(space.Applications[2].Name).To(Equal("app-3"))
		})

		It("has service instances", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
				ServiceInstances: []models.ServiceInstanceFields{
					{Guid: "service-1-guid", Name: "service-1"},
					{Guid: "service-2-guid", Name: "service-2"},
				},
			}

			Expect(len(space.ServiceInstances)).To(Equal(2))
			Expect(space.ServiceInstances[0].Name).To(Equal("service-1"))
			Expect(space.ServiceInstances[1].Name).To(Equal("service-2"))
		})

		It("has domains", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
				Domains: []models.DomainFields{
					{Guid: "domain-1-guid", Name: "example.com"},
					{Guid: "domain-2-guid", Name: "test.com"},
				},
			}

			Expect(len(space.Domains)).To(Equal(2))
			Expect(space.Domains[0].Name).To(Equal("example.com"))
			Expect(space.Domains[1].Name).To(Equal("test.com"))
		})

		It("has security groups", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
				SecurityGroups: []models.SecurityGroupFields{
					{Guid: "sg-1-guid", Name: "sg-1"},
					{Guid: "sg-2-guid", Name: "sg-2"},
				},
			}

			Expect(len(space.SecurityGroups)).To(Equal(2))
			Expect(space.SecurityGroups[0].Name).To(Equal("sg-1"))
			Expect(space.SecurityGroups[1].Name).To(Equal("sg-2"))
		})

		It("has space quota guid", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
				SpaceQuotaGuid: "space-quota-guid",
			}

			Expect(space.SpaceQuotaGuid).To(Equal("space-quota-guid"))
		})

		It("can have empty collections", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
				Applications:     []models.ApplicationFields{},
				ServiceInstances: []models.ServiceInstanceFields{},
				Domains:          []models.DomainFields{},
				SecurityGroups:   []models.SecurityGroupFields{},
			}

			Expect(len(space.Applications)).To(Equal(0))
			Expect(len(space.ServiceInstances)).To(Equal(0))
			Expect(len(space.Domains)).To(Equal(0))
			Expect(len(space.SecurityGroups)).To(Equal(0))
		})

		It("can have empty space quota guid", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
				SpaceQuotaGuid: "",
			}

			Expect(space.SpaceQuotaGuid).To(BeEmpty())
		})
	})
})
