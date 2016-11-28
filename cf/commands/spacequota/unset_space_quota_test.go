package spacequota_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/spacequotas/spacequotasfakes"
	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
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

var _ = Describe("unset-space-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *spacequotasfakes.FakeSpaceQuotaRepository
		spaceRepo           *spacesfakes.FakeSpaceRepository
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("unset-space-quota").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		quotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("unset-space-quota", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("fails with usage when provided too many or two few args", func() {
		runCommand("space")
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))

		runCommand("space", "quota", "extra-stuff")
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
	})

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand("space", "quota")).To(BeFalse())
		})

		It("requires the user to target an org", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			orgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			orgReq.ExecuteReturns(errors.New("not targeting org"))
			requirementsFactory.NewTargetedOrgRequirementReturns(orgReq)

			Expect(runCommand("space", "quota")).To(BeFalse())
		})
	})

	Context("when requirements are met", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
		})

		It("unassigns a quota from a space", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Name: "my-space",
					GUID: "my-space-guid",
				},
			}

			quota := models.SpaceQuota{Name: "my-quota", GUID: "my-quota-guid"}

			quotaRepo.FindByNameReturns(quota, nil)
			spaceRepo.FindByNameReturns(space, nil)

			runCommand("my-space", "my-quota")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Unassigning space quota", "my-quota", "my-space", "my-user"},
				[]string{"OK"},
			))

			Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("my-quota"))
			spaceGUID, quotaGUID := quotaRepo.UnassignQuotaFromSpaceArgsForCall(0)
			Expect(spaceGUID).To(Equal("my-space-guid"))
			Expect(quotaGUID).To(Equal("my-quota-guid"))
		})
	})
})
