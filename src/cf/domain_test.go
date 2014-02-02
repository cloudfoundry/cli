package cf

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestRouteURL", func() {

			route := Route{}
			route.Host = "foo"

			domain := DomainFields{}
			domain.Name = "example.com"
			route.Domain = domain

			assert.Equal(mr.T(), route.URL(), "foo.example.com")
		})
		It("TestRouteURLWithoutHost", func() {

			route := Route{}
			route.Host = ""

			domain := DomainFields{}
			domain.Name = "example.com"
			route.Domain = domain

			assert.Equal(mr.T(), route.URL(), "example.com")
		})
	})
}
