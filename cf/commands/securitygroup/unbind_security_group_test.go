package securitygroup_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/api/securitygroups/securitygroupsfakes"
	"code.cloudfoundry.org/cli/cf/api/securitygroups/spaces/spacesfakes"
	spacesapifakes "code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("unbind-security-group command", func() {
	var (
		ui                  *testterm.FakeUI
		securityGroupRepo   *securitygroupsfakes.FakeSecurityGroupRepo
		orgRepo             *organizationsfakes.FakeOrganizationRepository
		spaceRepo           *spacesapifakes.FakeSpaceRepository
		secBinder           *spacesfakes.FakeSecurityGroupSpaceBinder
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupRepository(securityGroupRepo)
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupSpaceBinder(secBinder)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("unbind-security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		securityGroupRepo = new(securitygroupsfakes.FakeSecurityGroupRepo)
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		spaceRepo = new(spacesapifakes.FakeSpaceRepository)
		secBinder = new(spacesfakes.FakeSecurityGroupSpaceBinder)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("unbind-security-group", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("should fail if not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-group")).To(BeFalse())
		})

		It("should fail with usage when not provided with any arguments", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("should fail with usage when provided with a number of arguments that is either 2 or 4 or a number larger than 4", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand("I", "like")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
			runCommand("Turn", "down", "for", "what")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
			runCommand("My", "Very", "Excellent", "Mother", "Just", "Sat", "Under", "Nine", "ThingsThatArentPlanets")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		Context("when everything exists", func() {
			BeforeEach(func() {
				securityGroup := models.SecurityGroup{
					SecurityGroupFields: models.SecurityGroupFields{
						Name:  "my-group",
						GUID:  "my-group-guid",
						Rules: []map[string]interface{}{},
					},
				}

				securityGroupRepo.ReadReturns(securityGroup, nil)

				orgRepo.ListOrgsReturns([]models.Organization{{
					OrganizationFields: models.OrganizationFields{
						Name: "my-org",
						GUID: "my-org-guid",
					}},
				}, nil)

				space := models.Space{SpaceFields: models.SpaceFields{Name: "my-space", GUID: "my-space-guid"}}
				spaceRepo.FindByNameInOrgReturns(space, nil)
			})

			It("removes the security group when we only pass the security group name (using the targeted org and space)", func() {
				runCommand("my-group")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Unbinding security group", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				securityGroupGUID, spaceGUID := secBinder.UnbindSpaceArgsForCall(0)
				Expect(securityGroupGUID).To(Equal("my-group-guid"))
				Expect(spaceGUID).To(Equal("my-space-guid"))
			})

			It("removes the security group when we pass the org and space", func() {
				runCommand("my-group", "my-org", "my-space")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Unbinding security group", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				securityGroupGUID, spaceGUID := secBinder.UnbindSpaceArgsForCall(0)
				Expect(securityGroupGUID).To(Equal("my-group-guid"))
				Expect(spaceGUID).To(Equal("my-space-guid"))
			})
		})

		Context("when one of the things does not exist", func() {
			BeforeEach(func() {
				securityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.New("I accidentally the"))
			})

			It("fails with an error", func() {
				runCommand("my-group", "my-org", "my-space")
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
			})
		})
	})
})
