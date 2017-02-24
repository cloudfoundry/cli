package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpaceRole", func() {
	var spaceRole SpaceRole

	Describe("Complete", func() {
		DescribeTable("returns list of completions",
			func(prefix string, matches []flags.Completion) {
				completions := spaceRole.Complete(prefix)
				Expect(completions).To(Equal(matches))
			},
			Entry("returns 'SpaceManager', 'SpaceDeveloper' and 'SpaceAuditor' when passed 'S'", "S",
				[]flags.Completion{{Item: "SpaceManager"}, {Item: "SpaceDeveloper"}, {Item: "SpaceAuditor"}}),
			Entry("returns 'SpaceManager', 'SpaceDeveloper' and 'SpaceAuditor' when passed 's'", "s",
				[]flags.Completion{{Item: "SpaceManager"}, {Item: "SpaceDeveloper"}, {Item: "SpaceAuditor"}}),
			Entry("completes to 'SpaceAuditor' when passed 'Spacea'", "Spacea",
				[]flags.Completion{{Item: "SpaceAuditor"}}),
			Entry("completes to 'SpaceDeveloper' when passed 'Spaced'", "Spaced",
				[]flags.Completion{{Item: "SpaceDeveloper"}}),
			Entry("completes to 'SpaceManager' when passed 'Spacem'", "Spacem",
				[]flags.Completion{{Item: "SpaceManager"}}),
			Entry("completes to 'SpaceManager' when passed 'spacEM'", "spacEM",
				[]flags.Completion{{Item: "SpaceManager"}}),
			Entry("returns 'SpaceManager', 'SpaceDeveloper' and 'SpaceAuditor' when passed nothing", "",
				[]flags.Completion{{Item: "SpaceManager"}, {Item: "SpaceDeveloper"}, {Item: "SpaceAuditor"}}),
			Entry("completes to nothing when passed 'wut'", "wut",
				[]flags.Completion{}),
		)
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			spaceRole = SpaceRole{}
		})

		It("accepts SpaceManager", func() {
			err := spaceRole.UnmarshalFlag("spacemanager")
			Expect(err).ToNot(HaveOccurred())
			Expect(spaceRole).To(Equal(SpaceRole{Role: "SpaceManager"}))
		})

		It("accepts SpaceDeveloper", func() {
			err := spaceRole.UnmarshalFlag("Spacedeveloper")
			Expect(err).ToNot(HaveOccurred())
			Expect(spaceRole).To(Equal(SpaceRole{Role: "SpaceDeveloper"}))
		})

		It("accepts SpaceAuditor", func() {
			err := spaceRole.UnmarshalFlag("spaceAuditor")
			Expect(err).ToNot(HaveOccurred())
			Expect(spaceRole).To(Equal(SpaceRole{Role: "SpaceAuditor"}))
		})

		It("errors on anything else", func() {
			err := spaceRole.UnmarshalFlag("I AM A BANANANANANANANANA")
			Expect(err).To(MatchError(&flags.Error{
				Type:    flags.ErrRequired,
				Message: `ROLE must be "SpaceManager", "SpaceDeveloper" and "SpaceAuditor"`,
			}))
			Expect(spaceRole.Role).To(BeEmpty())
		})
	})
})
