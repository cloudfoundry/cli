package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworkPort", func() {
	var port NetworkPort

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			port = NetworkPort{}
		})

		DescribeTable("it sets the ports correctly",
			func(input string, expectedStart int, expectedEnd int) {
				err := port.UnmarshalFlag(input)
				Expect(err).ToNot(HaveOccurred())
				Expect(port).To(Equal(NetworkPort{
					StartPort: expectedStart,
					EndPort:   expectedEnd,
				}))
			},
			Entry("when provided '3000' it sets the start and end to 3000", "3000", 3000, 3000),
			Entry("when provided '3000-6000' it sets the start to 3000 and end to 6000", "3000-6000", 3000, 6000),
		)

		DescribeTable("errors correctly",
			func(input string, expectedErr error) {
				err := port.UnmarshalFlag(input)
				Expect(err).To(MatchError(expectedErr))
			},

			Entry("when provided 'fooo' it returns back a flag error", "fooo",
				&flags.Error{
					Type:    flags.ErrRequired,
					Message: `PORT must be a positive integer`,
				}),
			Entry("when provided '1-fooo' it returns back a flag error", "1-fooo",
				&flags.Error{
					Type:    flags.ErrRequired,
					Message: `PORT must be a positive integer`,
				}),
			Entry("when provided '-1' it returns back a flag error", "-1",
				&flags.Error{
					Type:    flags.ErrRequired,
					Message: `PORT must be a positive integer`,
				}),
			Entry("when provided '-1-1' it returns back a flag error", "-1-1",
				&flags.Error{
					Type:    flags.ErrRequired,
					Message: `PORT must be a positive integer`,
				}),
			Entry("when provided '1-2-3' it returns back a flag error", "1-2-3",
				&flags.Error{
					Type:    flags.ErrRequired,
					Message: `PORT syntax must match integer[-integer]`,
				}),
		)
	})
})
