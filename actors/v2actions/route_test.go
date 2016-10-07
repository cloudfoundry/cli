package v2actions_test

import (
	. "code.cloudfoundry.org/cli/actors/v2actions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actions", func() {
	Describe("Route", func() {
		DescribeTable("String", func(hostname string, domain string, path string, port int, expectedValue string) {
			route := Route{
				Hostname: hostname,
				Domain:   domain,
				Path:     path,
				Port:     port,
			}
			actualValue := route.String()

			Expect(actualValue).To(Equal(expectedValue))
		},

			Entry("has domain", "", "domain.com", "", 0, "domain.com"),
			Entry("has hostname, domain", "host", "domain.com", "", 0, "host.domain.com"),
			Entry("has hostname, domain, path", "host", "domain.com", "path", 0, "host.domain.com/path"),
			Entry("has domain, port", "", "domain.com", "", 3333, "domain.com:3333"),
		)
	})
})
