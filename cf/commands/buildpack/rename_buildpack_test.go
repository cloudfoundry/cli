package buildpack_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("rename-buildpack command", func() {
	var (
		fakeRepo            *apifakes.OldFakeBuildpackRepository
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetBuildpackRepository(fakeRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("rename-buildpack").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewBuildpackRequirementReturns(new(requirementsfakes.FakeBuildpackRequirement))
		ui = new(testterm.FakeUI)
		fakeRepo = new(apifakes.OldFakeBuildpackRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("rename-buildpack", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("fails requirements when called without the current name and the new name to use", func() {
		passed := runCommand("my-buildpack-name")
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
		Expect(passed).To(BeFalse())
	})

	Context("when logged in", func() {
		It("renames a buildpack", func() {
			fakeRepo.FindByNameBuildpack = models.Buildpack{
				Name: "my-buildpack",
				GUID: "my-buildpack-guid",
			}

			runCommand("my-buildpack", "new-buildpack")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"OK"},
			))
		})

		It("fails when the buildpack does not exist", func() {
			fakeRepo.FindByNameNotFound = true

			runCommand("my-buildpack1", "new-buildpack")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"FAILED"},
				[]string{"Buildpack my-buildpack1 not found"},
			))
		})

		It("fails when there is an error updating the buildpack", func() {
			fakeRepo.FindByNameBuildpack = models.Buildpack{
				Name: "my-buildpack",
				GUID: "my-buildpack-guid",
			}
			fakeRepo.UpdateBuildpackReturns.Error = errors.New("SAD TROMBONE")

			runCommand("my-buildpack1", "new-buildpack")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"SAD TROMBONE"},
			))
		})
	})
})
