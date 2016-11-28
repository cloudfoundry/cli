package securitygroup_test

import (
	"code.cloudfoundry.org/cli/cf/api/securitygroups/securitygroupsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-security-group command", func() {
	var (
		ui                  *testterm.FakeUI
		securityGroupRepo   *securitygroupsfakes.FakeSecurityGroupRepo
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupRepository(securityGroupRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		securityGroupRepo = new(securitygroupsfakes.FakeSecurityGroupRepo)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete-security-group", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("should fail if not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-group")).To(BeFalse())
		})

		It("should fail with usage when not provided a single argument", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand("whoops", "I can't believe", "I accidentally", "the whole thing")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		Context("when the group with the given name exists", func() {
			BeforeEach(func() {
				securityGroupRepo.ReadReturns(models.SecurityGroup{
					SecurityGroupFields: models.SecurityGroupFields{
						Name: "my-group",
						GUID: "group-guid",
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
				Expect(ui.Outputs()).To(ContainSubstrings(
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

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
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

			Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
		})
	})
})
