package securitygroup_test

import (
	"github.com/cloudfoundry/cli/cf/api/organizations/organizationsfakes"
	"github.com/cloudfoundry/cli/cf/api/securitygroups/securitygroupsfakes"
	"github.com/cloudfoundry/cli/cf/api/securitygroups/spaces/spacesfakes"
	spacesapifakes "github.com/cloudfoundry/cli/cf/api/spaces/spacesfakes"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bind-security-group command", func() {
	var (
		ui                    *testterm.FakeUI
		configRepo            coreconfig.Repository
		fakeSecurityGroupRepo *securitygroupsfakes.FakeSecurityGroupRepo
		requirementsFactory   *requirementsfakes.FakeFactory
		fakeSpaceRepo         *spacesapifakes.FakeSpaceRepository
		fakeOrgRepo           *organizationsfakes.FakeOrganizationRepository
		fakeSpaceBinder       *spacesfakes.FakeSecurityGroupSpaceBinder
		deps                  commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(fakeSpaceRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(fakeOrgRepo)
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupRepository(fakeSecurityGroupRepo)
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupSpaceBinder(fakeSpaceBinder)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("bind-security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		fakeOrgRepo = new(organizationsfakes.FakeOrganizationRepository)
		fakeSpaceRepo = new(spacesapifakes.FakeSpaceRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
		fakeSecurityGroupRepo = new(securitygroupsfakes.FakeSecurityGroupRepo)
		configRepo = testconfig.NewRepositoryWithDefaults()
		fakeSpaceBinder = new(spacesfakes.FakeSecurityGroupSpaceBinder)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("bind-security-group", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-craaaaaazy-security-group", "my-org", "my-space")).To(BeFalse())
		})

		It("succeeds when the user is logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			Expect(runCommand("my-craaaaaazy-security-group", "my-org", "my-space")).To(BeTrue())
		})

		It("fails with usage when not provided the name of a security group, org, and space", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand("one fish", "two fish", "three fish", "purple fish")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})
	})

	Context("when the user is logged in and provides the name of a security group", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		Context("when a security group with that name does not exist", func() {
			BeforeEach(func() {
				fakeSecurityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.NewModelNotFoundError("security group", "my-nonexistent-security-group"))
			})

			It("fails and tells the user", func() {
				runCommand("my-nonexistent-security-group", "my-org", "my-space")

				Expect(fakeSecurityGroupRepo.ReadArgsForCall(0)).To(Equal("my-nonexistent-security-group"))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"security group", "my-nonexistent-security-group", "not found"},
				))
			})
		})

		Context("when the org does not exist", func() {
			BeforeEach(func() {
				fakeOrgRepo.FindByNameReturns(models.Organization{}, errors.New("Org org not found"))
			})

			It("fails and tells the user", func() {
				runCommand("sec group", "org", "space")

				Expect(fakeOrgRepo.FindByNameArgsForCall(0)).To(Equal("org"))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Org", "org", "not found"},
				))
			})
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "org-name"
				org.GUID = "org-guid"
				fakeOrgRepo.ListOrgsReturns([]models.Organization{org}, nil)
				fakeOrgRepo.FindByNameReturns(org, nil)
				fakeSpaceRepo.FindByNameInOrgReturns(models.Space{}, errors.NewModelNotFoundError("Space", "space-name"))
			})

			It("fails and tells the user", func() {
				runCommand("sec group", "org-name", "space-name")

				name, orgGUID := fakeSpaceRepo.FindByNameInOrgArgsForCall(0)
				Expect(name).To(Equal("space-name"))
				Expect(orgGUID).To(Equal("org-guid"))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Space", "space-name", "not found"},
				))
			})
		})

		Context("everything is hunky dory", func() {
			BeforeEach(func() {
				org := models.Organization{}
				org.Name = "org-name"
				org.GUID = "org-guid"
				fakeOrgRepo.ListOrgsReturns([]models.Organization{org}, nil)

				space := models.Space{}
				space.Name = "space-name"
				space.GUID = "space-guid"
				fakeSpaceRepo.FindByNameInOrgReturns(space, nil)

				securityGroup := models.SecurityGroup{}
				securityGroup.Name = "security-group"
				securityGroup.GUID = "security-group-guid"
				fakeSecurityGroupRepo.ReadReturns(securityGroup, nil)
			})

			JustBeforeEach(func() {
				runCommand("security-group", "org-name", "space-name")
			})

			It("assigns the security group to the space", func() {
				secGroupGUID, spaceGUID := fakeSpaceBinder.BindSpaceArgsForCall(0)
				Expect(secGroupGUID).To(Equal("security-group-guid"))
				Expect(spaceGUID).To(Equal("space-guid"))
			})

			It("describes what it is doing for the user's benefit", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Assigning security group security-group to space space-name in org org-name as my-user"},
					[]string{"OK"},
					[]string{"TIP: Changes will not apply to existing running applications until they are restarted."},
				))
			})
		})
	})
})
