package flag_test

import (
	"fmt"

	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpaceRole", func() {
	var forward SSHPortForwarding

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			forward = SSHPortForwarding{}
		})

		Context("when passed local_port:remote:remote_port", func() {
			It("extracts the local and remote addresses", func() {
				err := forward.UnmarshalFlag("8888:remote:8080")
				Expect(err).ToNot(HaveOccurred())
				Expect(forward).To(Equal(SSHPortForwarding{
					LocalAddress:  "localhost:8888",
					RemoteAddress: "remote:8080",
				}))
			})
		})

		Context("when passed local:local_port:remote:remote_port", func() {
			It("extracts the local and remote addresses", func() {
				err := forward.UnmarshalFlag("local:8888:remote:8080")
				Expect(err).ToNot(HaveOccurred())
				Expect(forward).To(Equal(SSHPortForwarding{
					LocalAddress:  "local:8888",
					RemoteAddress: "remote:8080",
				}))
			})
		})

		DescribeTable("error cases",
			func(input string) {
				err := forward.UnmarshalFlag(input)
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: fmt.Sprintf("Bad local forwarding specification '%s'", input),
				}))
			},

			Entry("no colons", "IAMABANANA909009009"),
			Entry("1 colon", "IAMABANANA:909009009"),
			Entry("too many colons", "I:AM:A:BANANA:909009009"),
			Entry("empty values in between colons", "I:AM:A:"),
			Entry("[implicit localhost] incorrect port numbers for first value", "I:AM:8888"),
			Entry("[implicit localhost] incorrect port numbers for third value", "8888:AM:potato"),
			Entry("[explicit localhost] incorrect port numbers for second value", "localhost:foo:AM:8888"),
			Entry("[explicit localhost] incorrect port numbers for fourth value", "localhost:8080:AM:bar"),
		)
	})
})
