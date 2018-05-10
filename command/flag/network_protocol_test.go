package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworkProtocol", func() {
	var proto NetworkProtocol

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := proto.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},
			Entry("returns 'tcp' when passed 't'", "t",
				[]flags.Completion{{Item: "tcp"}}),
			Entry("returns 'tcp' when passed 'T'", "T",
				[]flags.Completion{{Item: "tcp"}}),
			Entry("returns 'udp' when passed 'u'", "u",
				[]flags.Completion{{Item: "udp"}}),
			Entry("returns 'udp' when passed 'U'", "U",
				[]flags.Completion{{Item: "udp"}}),
			Entry("returns 'tcp' and 'udp' when passed ''", "",
				[]flags.Completion{{Item: "tcp"}, {Item: "udp"}}),
		)
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			proto = NetworkProtocol{}
		})

		DescribeTable("downcases and sets type",
			func(input string, expectedProtocol string) {
				err := proto.UnmarshalFlag(input)
				Expect(err).ToNot(HaveOccurred())
				Expect(proto.Protocol).To(Equal(expectedProtocol))
			},
			Entry("sets 'tcp' when passed 'tcp'", "tcp", "tcp"),
			Entry("sets 'tcp' when passed 'tCp'", "tCp", "tcp"),
			Entry("sets 'udp' when passed 'udp'", "udp", "udp"),
			Entry("sets 'udp' when passed 'uDp'", "uDp", "udp"),
		)

		Context("when passed anything else", func() {
			It("returns an error", func() {
				err := proto.UnmarshalFlag("banana")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `PROTOCOL must be "tcp" or "udp"`,
				}))
				Expect(proto.Protocol).To(BeEmpty())
			})
		})
	})
})
