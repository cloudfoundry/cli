package spacequota_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("unset-space-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *fakes.FakeSpaceQuotaRepository
		spaceRepo           *testapi.FakeSpaceRepository
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("unset-space-quota").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("unset-space-quota", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails with usage when provided too many or two few args", func() {
		runCommand("space")
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))

		runCommand("space", "quota", "extra-stuff")
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
	})

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false

			Expect(runCommand("space", "quota")).To(BeFalse())
		})

		It("requires the user to target an org", func() {
			requirementsFactory.TargetedOrgSuccess = false

			Expect(runCommand("space", "quota")).To(BeFalse())
		})
	})

	Context("when requirements are met", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
		})

		It("unassigns a quota from a space", func() {
			space := models.Space{
				SpaceFields: models.SpaceFields{
					Name: "my-space",
					Guid: "my-space-guid",
				},
			}

			quota := models.SpaceQuota{Name: "my-quota", Guid: "my-quota-guid"}

			quotaRepo.FindByNameReturns(quota, nil)
			spaceRepo.FindByNameReturns(space, nil)

			runCommand("my-space", "my-quota")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Unassigning space quota", "my-quota", "my-space", "my-user"},
				[]string{"OK"},
			))

			Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("my-quota"))
			spaceGuid, quotaGuid := quotaRepo.UnassignQuotaFromSpaceArgsForCall(0)
			Expect(spaceGuid).To(Equal("my-space-guid"))
			Expect(quotaGuid).To(Equal("my-quota-guid"))
		})
	})
})
