package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Revision", func() {
	var revision Revision

	BeforeEach(func() {
		revision = Revision{}
	})

	Describe("UnmarshalFlag", func() {
		When("the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := revision.IsValidValue("")
				Expect(err).ToNot(HaveOccurred())
				Expect(revision).To(Equal(Revision{NullInt: types.NullInt{Value: 0, IsSet: false}}))
			})
		})

		When("an invalid integer is provided", func() {
			It("returns an error", func() {
				err := revision.IsValidValue("abcdef")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid argument for flag '--revision' (expected int > 0)",
				}))
				Expect(revision).To(Equal(Revision{NullInt: types.NullInt{Value: 0, IsSet: false}}))
			})
		})

		When("a negative integer is provided", func() {
			It("returns an error", func() {
				err := revision.IsValidValue("-10")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid argument for flag '--revision' (expected int > 0)",
				}))
				Expect(revision).To(Equal(Revision{NullInt: types.NullInt{Value: -10, IsSet: true}}))
			})
		})

		When("0 is provided", func() {
			It("returns an error", func() {
				err := revision.IsValidValue("0")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid argument for flag '--revision' (expected int > 0)",
				}))
				Expect(revision).To(Equal(Revision{NullInt: types.NullInt{Value: 0, IsSet: true}}))
			})
		})

		When("a valid integer is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := revision.IsValidValue("1")
				Expect(err).ToNot(HaveOccurred())
				Expect(revision).To(Equal(Revision{NullInt: types.NullInt{Value: 1, IsSet: true}}))
			})
		})
	})
})
