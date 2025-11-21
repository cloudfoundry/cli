package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain Additional Tests", func() {
	Describe("DomainFields", func() {
		Describe("UrlForHost", func() {
			It("returns full URL with host and domain", func() {
				domain := models.DomainFields{
					Guid: "domain-guid",
					Name: "example.com",
				}

				url := domain.UrlForHost("my-app")
				Expect(url).To(Equal("my-app.example.com"))
			})

			It("returns only domain when host is empty", func() {
				domain := models.DomainFields{
					Guid: "domain-guid",
					Name: "example.com",
				}

				url := domain.UrlForHost("")
				Expect(url).To(Equal("example.com"))
			})

			It("handles subdomain in host", func() {
				domain := models.DomainFields{
					Name: "example.com",
				}

				url := domain.UrlForHost("app.subdomain")
				Expect(url).To(Equal("app.subdomain.example.com"))
			})

			It("handles different domains", func() {
				domain1 := models.DomainFields{Name: "example.com"}
				domain2 := models.DomainFields{Name: "test.io"}
				domain3 := models.DomainFields{Name: "mydomain.org"}

				Expect(domain1.UrlForHost("api")).To(Equal("api.example.com"))
				Expect(domain2.UrlForHost("www")).To(Equal("www.test.io"))
				Expect(domain3.UrlForHost("app")).To(Equal("app.mydomain.org"))
			})

			It("handles hosts with dashes", func() {
				domain := models.DomainFields{
					Name: "example.com",
				}

				url := domain.UrlForHost("my-app-name")
				Expect(url).To(Equal("my-app-name.example.com"))
			})
		})

		It("stores domain fields", func() {
			domain := models.DomainFields{
				Guid:                   "domain-guid",
				Name:                   "example.com",
				OwningOrganizationGuid: "org-guid",
				Shared:                 true,
			}

			Expect(domain.Guid).To(Equal("domain-guid"))
			Expect(domain.Name).To(Equal("example.com"))
			Expect(domain.OwningOrganizationGuid).To(Equal("org-guid"))
			Expect(domain.Shared).To(BeTrue())
		})

		It("handles private domain", func() {
			domain := models.DomainFields{
				Guid:                   "domain-guid",
				Name:                   "private.example.com",
				OwningOrganizationGuid: "org-guid",
				Shared:                 false,
			}

			Expect(domain.Shared).To(BeFalse())
			Expect(domain.OwningOrganizationGuid).ToNot(BeEmpty())
		})

		It("handles shared domain", func() {
			domain := models.DomainFields{
				Guid:                   "domain-guid",
				Name:                   "shared.example.com",
				OwningOrganizationGuid: "",
				Shared:                 true,
			}

			Expect(domain.Shared).To(BeTrue())
			Expect(domain.OwningOrganizationGuid).To(BeEmpty())
		})

		It("handles empty values", func() {
			domain := models.DomainFields{}

			Expect(domain.Guid).To(BeEmpty())
			Expect(domain.Name).To(BeEmpty())
			Expect(domain.OwningOrganizationGuid).To(BeEmpty())
			Expect(domain.Shared).To(BeFalse())
		})

		It("handles different domain name formats", func() {
			domain1 := models.DomainFields{Name: "example.com"}
			domain2 := models.DomainFields{Name: "sub.example.com"}
			domain3 := models.DomainFields{Name: "cf.example.io"}
			domain4 := models.DomainFields{Name: "app-domain.org"}

			Expect(domain1.Name).To(Equal("example.com"))
			Expect(domain2.Name).To(Equal("sub.example.com"))
			Expect(domain3.Name).To(Equal("cf.example.io"))
			Expect(domain4.Name).To(Equal("app-domain.org"))
		})
	})
})
