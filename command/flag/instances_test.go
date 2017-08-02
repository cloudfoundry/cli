package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instances", func() {
	var instances Instances

	BeforeEach(func() {
		instances = Instances{}
	})

	Describe("UnmarshalFlag", func() {
		Context("when the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := instances.UnmarshalFlag("")
				Expect(err).ToNot(HaveOccurred())
				Expect(instances).To(Equal(Instances{NullInt: types.NullInt{Value: 0, IsSet: false}}))
			})
		})

		Context("when an invalid integer is provided", func() {
			It("returns an error", func() {
				err := instances.UnmarshalFlag("abcdef")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid argument for flag '-i' (expected int > 0)",
				}))
				Expect(instances).To(Equal(Instances{NullInt: types.NullInt{Value: 0, IsSet: false}}))
			})
		})

		Context("when a negative integer is provided", func() {
			It("returns an error", func() {
				err := instances.UnmarshalFlag("-10")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid argument for flag '-i' (expected int > 0)",
				}))
				Expect(instances).To(Equal(Instances{NullInt: types.NullInt{Value: -10, IsSet: true}}))
			})
		})

		Context("when a valid integer is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := instances.UnmarshalFlag("0")
				Expect(err).ToNot(HaveOccurred())
				Expect(instances).To(Equal(Instances{NullInt: types.NullInt{Value: 0, IsSet: true}}))
			})
		})
	})
})
