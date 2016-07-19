package models_test

import (
	"code.cloudfoundry.org/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteSummary", func() {
	Describe("URL", func() {
		var (
			r    models.RouteSummary
			host string
			path string
			port int
		)

		BeforeEach(func() {
			host = ""
			path = ""
			port = 0
		})

		JustBeforeEach(func() {
			r = models.RouteSummary{
				Host: host,
				Domain: models.DomainFields{
					Name: "the-domain",
				},
				Path: path,
				Port: port,
			}
		})

		Context("when the host is blank", func() {
			BeforeEach(func() {
				host = ""
			})

			It("returns the domain", func() {
				Expect(r.URL()).To(Equal("the-domain"))
			})

			Context("when the path is present", func() {
				BeforeEach(func() {
					path = "/the-path"
				})

				It("returns the domain and path", func() {
					Expect(r.URL()).To(Equal("the-domain/the-path"))
				})
			})

			Context("when the port is present", func() {
				BeforeEach(func() {
					port = 9001
				})

				It("returns the port", func() {
					Expect(r.URL()).To(Equal("the-domain:9001"))
				})
			})
		})

		Context("when the host is not blank", func() {
			BeforeEach(func() {
				host = "the-host"
			})

			It("returns the host and domain", func() {
				Expect(r.URL()).To(Equal("the-host.the-domain"))
			})

			Context("when the path is present", func() {
				BeforeEach(func() {
					path = "/the-path"
				})

				It("returns the host and domain and path", func() {
					Expect(r.URL()).To(Equal("the-host.the-domain/the-path"))
				})
			})
		})
	})
})
