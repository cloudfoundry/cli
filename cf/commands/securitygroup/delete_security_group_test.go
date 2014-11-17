package securitygroup_test

import (
	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-security-group command", func() {
	var (
		ui                  *testterm.FakeUI
		securityGroupRepo   *fakeSecurityGroup.FakeSecurityGroupRepo
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		securityGroupRepo = &fakeSecurityGroup.FakeSecurityGroupRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		cmd := NewDeleteSecurityGroup(ui, configRepo, securityGroupRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("should fail if not logged in", func() {
			Expect(runCommand("my-group")).To(BeFalse())
		})

		It("should fail with usage when not provided a single argument", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("whoops", "I can't believe", "I accidentally", "the whole thing")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when the group with the given name exists", func() {
			BeforeEach(func() {
				securityGroupRepo.ReadReturns(models.SecurityGroup{
					SecurityGroupFields: models.SecurityGroupFields{
						Name: "my-group",
						Guid: "group-guid",
					},
				}, nil)
			})

			Context("delete a security group", func() {
				It("when passed the -f flag", func() {
					runCommand("-f", "my-group")
					Expect(securityGroupRepo.ReadArgsForCall(0)).To(Equal("my-group"))
					Expect(securityGroupRepo.DeleteArgsForCall(0)).To(Equal("group-guid"))

					Expect(ui.Prompts).To(BeEmpty())
				})

				It("should prompt user when -f flag is not present", func() {
					ui.Inputs = []string{"y"}

					runCommand("my-group")
					Expect(securityGroupRepo.ReadArgsForCall(0)).To(Equal("my-group"))
					Expect(securityGroupRepo.DeleteArgsForCall(0)).To(Equal("group-guid"))

					Expect(ui.Prompts).To(ContainSubstrings(
						[]string{"Really delete the security group", "my-group"},
					))
				})

				It("should not delete when user passes 'n' to prompt", func() {
					ui.Inputs = []string{"n"}

					runCommand("my-group")
					Expect(securityGroupRepo.ReadCallCount()).To(Equal(0))
					Expect(securityGroupRepo.DeleteCallCount()).To(Equal(0))

					Expect(ui.Prompts).To(ContainSubstrings(
						[]string{"Really delete the security group", "my-group"},
					))
				})
			})

			It("tells the user what it's about to do", func() {
				runCommand("-f", "my-group")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "security group", "my-group", "my-user"},
					[]string{"OK"},
				))
			})
		})

		Context("when finding the group returns an error", func() {
			BeforeEach(func() {
				securityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.New("pbbbbbbbbbbt"))
			})

			It("fails and tells the user", func() {
				runCommand("-f", "whoops")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when a group with that name does not exist", func() {
			BeforeEach(func() {
				securityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.NewModelNotFoundError("Security group", "uh uh uh -- you didn't sahy the magick word"))
			})

			It("fails and tells the user", func() {
				runCommand("-f", "whoop")

				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"whoop", "does not exist"}))
			})
		})

		It("fails and warns the user if deleting fails", func() {
			securityGroupRepo.DeleteReturns(errors.New("raspberry"))
			runCommand("-f", "whoops")

			Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
		})
	})
})
