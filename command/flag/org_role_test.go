package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrgRole", func() {
	var orgRole OrgRole

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := orgRole.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},
			Entry("returns 'OrgManager' and 'OrgAuditor' when passed 'O'", "O",
				[]flags.Completion{{Item: "OrgManager"}, {Item: "OrgAuditor"}}),
			Entry("returns 'OrgManager' and 'OrgAuditor' when passed 'o'", "o",
				[]flags.Completion{{Item: "OrgManager"}, {Item: "OrgAuditor"}}),
			Entry("returns 'BillingManager' when passed 'B'", "B",
				[]flags.Completion{{Item: "BillingManager"}}),
			Entry("returns 'BillingManager' when passed 'b'", "b",
				[]flags.Completion{{Item: "BillingManager"}}),
			Entry("completes to 'OrgAuditor' when passed 'orgA'", "orgA",
				[]flags.Completion{{Item: "OrgAuditor"}}),
			Entry("completes to 'OrgManager' when passed 'orgm'", "orgm",
				[]flags.Completion{{Item: "OrgManager"}}),
			Entry("returns 'OrgManager', 'BillingManager' and 'OrgAuditor' when passed nothing", "",
				[]flags.Completion{{Item: "OrgManager"}, {Item: "BillingManager"}, {Item: "OrgAuditor"}}),
			Entry("completes to nothing when passed 'wut'", "wut",
				[]flags.Completion{}),
		)
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			orgRole = OrgRole{}
		})

		It("accepts OrgManager", func() {
			err := orgRole.UnmarshalFlag("orgmanager")
			Expect(err).ToNot(HaveOccurred())
			Expect(orgRole).To(Equal(OrgRole{Role: "OrgManager"}))
		})

		It("accepts BillingManager", func() {
			err := orgRole.UnmarshalFlag("Billingmanager")
			Expect(err).ToNot(HaveOccurred())
			Expect(orgRole).To(Equal(OrgRole{Role: "BillingManager"}))
		})

		It("accepts OrgAuditor", func() {
			err := orgRole.UnmarshalFlag("orgAuditor")
			Expect(err).ToNot(HaveOccurred())
			Expect(orgRole).To(Equal(OrgRole{Role: "OrgAuditor"}))
		})

		It("errors on anything else", func() {
			err := orgRole.UnmarshalFlag("I AM A BANANANANANANANANA")
			Expect(err).To(MatchError(&flags.Error{
				Type:    flags.ErrRequired,
				Message: `ROLE must be "OrgManager", "BillingManager" and "OrgAuditor"`,
			}))
			Expect(orgRole.Role).To(BeEmpty())
		})
	})
})
