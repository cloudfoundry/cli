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

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			orgRole = OrgRole{}
		})

		It("accepts OrgManager", func() {
			err := orgRole.UnmarshalFlag("orgmanager")
			Expect(err).ToNot(HaveOccurred())
			Expect(orgRole).To(Equal(OrgRole{Role: "OrgManager"}))
		})

		It("accepts OrgDeveloper", func() {
			err := orgRole.UnmarshalFlag("Orgdeveloper")
			Expect(err).ToNot(HaveOccurred())
			Expect(orgRole).To(Equal(OrgRole{Role: "OrgDeveloper"}))
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
				Message: `ROLE must be "OrgManager", "OrgDeveloper" and "OrgAuditor"`,
			}))
			Expect(orgRole.Role).To(BeEmpty())
		})
	})

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := orgRole.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},

			Entry("returns 'OrgManager', 'OrgDeveloper' and 'OrgAuditor' when passed 'O'", "O",
				[]flags.Completion{{Item: "OrgManager"}, {Item: "OrgDeveloper"}, {Item: "OrgAuditor"}}),
			Entry("returns 'OrgManager', 'OrgDeveloper' and 'OrgAuditor' when passed 'o'", "o",
				[]flags.Completion{{Item: "OrgManager"}, {Item: "OrgDeveloper"}, {Item: "OrgAuditor"}}),
			Entry("completes to 'OrgAuditor' when passed 'Orga'", "Orga",
				[]flags.Completion{{Item: "OrgAuditor"}}),
			Entry("completes to 'OrgDeveloper' when passed 'Orgd'", "Orgd",
				[]flags.Completion{{Item: "OrgDeveloper"}}),
			Entry("completes to 'OrgManager' when passed 'Orgm'", "Orgm",
				[]flags.Completion{{Item: "OrgManager"}}),
			Entry("returns 'OrgManager', 'OrgDeveloper' and 'OrgAuditor' when passed nothing", "",
				[]flags.Completion{{Item: "OrgManager"}, {Item: "OrgDeveloper"}, {Item: "OrgAuditor"}}),
			Entry("completes to nothing when passed 'wut'", "wut",
				[]flags.Completion{}),
		)
	})
})
