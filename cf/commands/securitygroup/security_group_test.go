package securitygroup_test

import (
	"code.cloudfoundry.org/cli/cf/api/securitygroups/securitygroupsfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("security-group command", func() {
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
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		securityGroupRepo = new(securitygroupsfakes.FakeSecurityGroupRepo)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("security-group", args, requirementsFactory, updateCommandDependency, false, ui)
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
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		Context("when the group with the given name exists", func() {
			BeforeEach(func() {
				rulesMap := []map[string]interface{}{{"just-pretend": "that-this-is-correct"}}
				securityGroup := models.SecurityGroup{
					SecurityGroupFields: models.SecurityGroupFields{
						Name:  "my-group",
						GUID:  "group-guid",
						Rules: rulesMap,
					},
					Spaces: []models.Space{
						{
							SpaceFields:  models.SpaceFields{GUID: "my-space-guid-1", Name: "space-1"},
							Organization: models.OrganizationFields{GUID: "my-org-guid-1", Name: "org-1"},
						},
						{
							SpaceFields:  models.SpaceFields{GUID: "my-space-guid", Name: "space-2"},
							Organization: models.OrganizationFields{GUID: "my-org-guid-1", Name: "org-2"},
						},
					},
				}

				securityGroupRepo.ReadReturns(securityGroup, nil)
			})

			It("should fetch the security group from its repo", func() {
				runCommand("my-group")
				Expect(securityGroupRepo.ReadArgsForCall(0)).To(Equal("my-group"))
			})

			It("tells the user what it's about to do and then shows the group", func() {
				runCommand("my-group")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting", "security group", "my-group", "my-user"},
					[]string{"OK"},
					[]string{"Name", "my-group"},
					[]string{"Rules"},
					[]string{"["},
					[]string{"{"},
					[]string{"just-pretend", "that-this-is-correct"},
					[]string{"}"},
					[]string{"]"},
					[]string{"#0", "org-1", "space-1"},
					[]string{"#1", "org-2", "space-2"},
				))
			})

			It("tells the user if no spaces are assigned", func() {
				securityGroup := models.SecurityGroup{
					SecurityGroupFields: models.SecurityGroupFields{
						Name:  "my-group",
						GUID:  "group-guid",
						Rules: []map[string]interface{}{},
					},
					Spaces: []models.Space{},
				}

				securityGroupRepo.ReadReturns(securityGroup, nil)

				runCommand("my-group")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"No spaces assigned"},
				))
			})
		})

		It("fails and warns the user if a group with that name could not be found", func() {
			securityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.New("half-past-tea-time"))
			runCommand("im-late!")

			Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
		})
	})
})
