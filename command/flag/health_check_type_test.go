package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"

	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("HealthCheckType", func() {
	var healthCheck HealthCheckType

	BeforeEach(func() {
		healthCheck = HealthCheckType{}
	})

	Describe("UnmarshalFlag", func() {
		DescribeTable("downcases and sets type",
			func(settingType string, expectedType string) {
				err := healthCheck.UnmarshalFlag(settingType)
				Expect(err).ToNot(HaveOccurred())
				Expect(healthCheck.Type).To(Equal(expectedType))
			},
			Entry("sets 'port' when passed 'port'", "port", "port"),
			Entry("sets 'port' when passed 'pOrt'", "pOrt", "port"),
			Entry("sets 'process' when passed 'none'", "none", "process"),
			Entry("sets 'process' when passed 'process'", "process", "process"),
			Entry("sets 'http' when passed 'http'", "http", "http"),
		)

		Context("when passed anything else", func() {
			It("returns an error", func() {
				err := healthCheck.UnmarshalFlag("banana")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `HEALTH_CHECK_TYPE must be "port", "process", or "http"`,
				}))
				Expect(healthCheck.Type).To(BeEmpty())
			})
		})
	})
})
