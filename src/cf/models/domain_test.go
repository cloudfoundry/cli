package models_test

import (
	. "cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with ginkgo", func() {
	var route Route
	BeforeEach(func() {
		route = Route{}

		domain := DomainFields{}
		domain.Name = "example.com"
		route.Domain = domain
	})

	It("TestRouteURL", func() {
		route.Host = "foo"
		Expect(route.URL()).To(Equal("foo.example.com"))
	})

	It("TestRouteURLWithoutHost", func() {
		route.Host = ""
		Expect(route.URL()).To(Equal("example.com"))
	})
})
