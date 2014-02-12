package models_test

import (
	. "cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
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
		assert.Equal(mr.T(), route.URL(), "foo.example.com")
	})

	It("TestRouteURLWithoutHost", func() {
		route.Host = ""
		assert.Equal(mr.T(), route.URL(), "example.com")
	})
})
