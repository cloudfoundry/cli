package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route", func() {
	Describe("Route", func() {
		Describe("URL", func() {
			It("returns full URL with host and domain", func() {
				route := models.Route{
					Guid: "route-guid",
					Host: "my-app",
					Domain: models.DomainFields{
						Name: "example.com",
					},
				}

				Expect(route.URL()).To(Equal("my-app.example.com"))
			})

			It("returns only domain when host is empty", func() {
				route := models.Route{
					Guid: "route-guid",
					Host: "",
					Domain: models.DomainFields{
						Name: "example.com",
					},
				}

				Expect(route.URL()).To(Equal("example.com"))
			})

			It("handles different domains", func() {
				route := models.Route{
					Host: "api",
					Domain: models.DomainFields{
						Name: "mydomain.io",
					},
				}

				Expect(route.URL()).To(Equal("api.mydomain.io"))
			})

			It("handles subdomain in host", func() {
				route := models.Route{
					Host: "app.subdomain",
					Domain: models.DomainFields{
						Name: "example.com",
					},
				}

				Expect(route.URL()).To(Equal("app.subdomain.example.com"))
			})
		})

		It("stores route fields", func() {
			route := models.Route{
				Guid: "route-guid",
				Host: "my-host",
				Domain: models.DomainFields{
					Guid: "domain-guid",
					Name: "example.com",
				},
				Space: models.SpaceFields{
					Guid: "space-guid",
					Name: "my-space",
				},
			}

			Expect(route.Guid).To(Equal("route-guid"))
			Expect(route.Host).To(Equal("my-host"))
			Expect(route.Domain.Guid).To(Equal("domain-guid"))
			Expect(route.Space.Guid).To(Equal("space-guid"))
		})

		It("has associated apps", func() {
			route := models.Route{
				Guid: "route-guid",
				Host: "my-host",
				Domain: models.DomainFields{
					Name: "example.com",
				},
				Apps: []models.ApplicationFields{
					{Guid: "app-1-guid", Name: "app-1"},
					{Guid: "app-2-guid", Name: "app-2"},
				},
			}

			Expect(len(route.Apps)).To(Equal(2))
			Expect(route.Apps[0].Name).To(Equal("app-1"))
			Expect(route.Apps[1].Name).To(Equal("app-2"))
		})

		It("can have no apps", func() {
			route := models.Route{
				Guid: "route-guid",
				Host: "my-host",
				Domain: models.DomainFields{
					Name: "example.com",
				},
				Apps: []models.ApplicationFields{},
			}

			Expect(len(route.Apps)).To(Equal(0))
		})
	})

	Describe("RouteSummary", func() {
		Describe("URL", func() {
			It("returns full URL with host and domain", func() {
				route := models.RouteSummary{
					Guid: "route-guid",
					Host: "my-app",
					Domain: models.DomainFields{
						Name: "example.com",
					},
				}

				Expect(route.URL()).To(Equal("my-app.example.com"))
			})

			It("returns only domain when host is empty", func() {
				route := models.RouteSummary{
					Guid: "route-guid",
					Host: "",
					Domain: models.DomainFields{
						Name: "example.com",
					},
				}

				Expect(route.URL()).To(Equal("example.com"))
			})

			It("handles different domains", func() {
				route := models.RouteSummary{
					Host: "www",
					Domain: models.DomainFields{
						Name: "mysite.org",
					},
				}

				Expect(route.URL()).To(Equal("www.mysite.org"))
			})

			It("formats like Route.URL", func() {
				fullRoute := models.Route{
					Host: "test",
					Domain: models.DomainFields{
						Name: "example.com",
					},
				}

				summaryRoute := models.RouteSummary{
					Host: "test",
					Domain: models.DomainFields{
						Name: "example.com",
					},
				}

				Expect(summaryRoute.URL()).To(Equal(fullRoute.URL()))
			})
		})

		It("stores route summary fields", func() {
			route := models.RouteSummary{
				Guid: "route-guid",
				Host: "my-host",
				Domain: models.DomainFields{
					Guid: "domain-guid",
					Name: "example.com",
				},
			}

			Expect(route.Guid).To(Equal("route-guid"))
			Expect(route.Host).To(Equal("my-host"))
			Expect(route.Domain.Name).To(Equal("example.com"))
		})
	})
})
