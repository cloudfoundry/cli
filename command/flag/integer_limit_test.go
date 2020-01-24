package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IntegerLimit", func() {
	var limit IntegerLimit

	BeforeEach(func() {
		limit = IntegerLimit{}
	})

	Describe("UnmarshalFlag", func() {
		When("the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := limit.UnmarshalFlag("")
				Expect(err).ToNot(HaveOccurred())
				Expect(limit).To(Equal(IntegerLimit{NullInt: types.NullInt{Value: 0, IsSet: false}}))
			})
		})

		When("an invalid value is provided", func() {
			It("returns an error", func() {
				err := limit.UnmarshalFlag("abcdef")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid integer limit (expected int >= -1)",
				}))
				Expect(limit).To(Equal(IntegerLimit{NullInt: types.NullInt{Value: 0, IsSet: false}}))
			})
		})

		When("an out of bounds integer is provided", func() {
			It("returns an error", func() {
				err := limit.UnmarshalFlag("-10")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid integer limit (expected int >= -1)",
				}))
				Expect(limit).To(Equal(IntegerLimit{NullInt: types.NullInt{Value: -10, IsSet: true}}))
			})
		})

		When("a valid integer is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := limit.UnmarshalFlag("-1")
				Expect(err).ToNot(HaveOccurred())
				Expect(limit).To(Equal(IntegerLimit{NullInt: types.NullInt{Value: -1, IsSet: true}}))
			})
		})
	})
})
