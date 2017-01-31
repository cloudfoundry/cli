package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"

	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HealthCheckType", func() {
	var healthCheck HealthCheckType

	BeforeEach(func() {
		healthCheck = HealthCheckType{}
	})

	Describe("UnmarshalFlag", func() {
		Context("when passed 'port'", func() {
			It("sets port to true", func() {
				err := healthCheck.UnmarshalFlag("port")
				Expect(err).ToNot(HaveOccurred())
				Expect(healthCheck.Type).To(Equal("port"))
			})
		})

		Context("when passed 'pOrt'", func() {
			It("sets port to true", func() {
				err := healthCheck.UnmarshalFlag("pOrt")
				Expect(err).ToNot(HaveOccurred())
				Expect(healthCheck.Type).To(Equal("port"))
			})
		})

		Context("when passed 'none'", func() {
			It("sets none to true", func() {
				err := healthCheck.UnmarshalFlag("none")
				Expect(err).ToNot(HaveOccurred())
				Expect(healthCheck.Type).To(Equal("none"))
			})
		})

		Context("when passed 'nOne'", func() {
			It("sets none to true", func() {
				err := healthCheck.UnmarshalFlag("nOne")
				Expect(err).ToNot(HaveOccurred())
				Expect(healthCheck.Type).To(Equal("none"))
			})
		})

		Context("when passed anything else", func() {
			It("returns an error", func() {
				err := healthCheck.UnmarshalFlag("banana")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `HEALTH_CHECK_TYPE must be "port" or "none"`,
				}))
				Expect(healthCheck.Type).To(BeEmpty())
			})
		})
	})
})
