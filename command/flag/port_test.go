package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Port", func() {
	var port Port
	BeforeEach(func() {
		port = Port{}
	})

	Describe("UnmarshalFlag", func() {
		Context("when the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := port.UnmarshalFlag("")
				Expect(err).ToNot(HaveOccurred())
				Expect(port).To(Equal(Port{NullInt: types.NullInt{Value: 0, IsSet: false}}))
			})
		})

		Context("when an invalid integer is provided", func() {
			It("returns an error", func() {
				err := port.UnmarshalFlag("abcdef")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid argument for flag '--port' (expected int > 0)",
				}))
				Expect(port).To(Equal(Port{NullInt: types.NullInt{Value: 0, IsSet: false}}))
			})
		})

		Context("when a negative integer is provided", func() {
			It("returns an error", func() {
				err := port.UnmarshalFlag("-10")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid argument for flag '--port' (expected int > 0)",
				}))
				Expect(port).To(Equal(Port{NullInt: types.NullInt{Value: -10, IsSet: true}}))
			})
		})

		Context("when a valid integer is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := port.UnmarshalFlag("0")
				Expect(err).ToNot(HaveOccurred())
				Expect(port).To(Equal(Port{NullInt: types.NullInt{Value: 0, IsSet: true}}))
			})
		})
	})
})
