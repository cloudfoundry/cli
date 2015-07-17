package buildpack_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("rename-buildpack command", func() {
	var (
		fakeRepo            *testapi.FakeBuildpackRepository
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetBuildpackRepository(fakeRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("rename-buildpack").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		ui = new(testterm.FakeUI)
		fakeRepo = &testapi.FakeBuildpackRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("rename-buildpack", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails requirements when called without the current name and the new name to use", func() {
		passed := runCommand("my-buildpack-name")
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
		Expect(passed).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("renames a buildpack", func() {
			fakeRepo.FindByNameBuildpack = models.Buildpack{
				Name: "my-buildpack",
				Guid: "my-buildpack-guid",
			}

			runCommand("my-buildpack", "new-buildpack")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"OK"},
			))
		})

		It("fails when the buildpack does not exist", func() {
			fakeRepo.FindByNameNotFound = true

			runCommand("my-buildpack1", "new-buildpack")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"FAILED"},
				[]string{"Buildpack my-buildpack1 not found"},
			))
		})

		It("fails when there is an error updating the buildpack", func() {
			fakeRepo.FindByNameBuildpack = models.Buildpack{
				Name: "my-buildpack",
				Guid: "my-buildpack-guid",
			}
			fakeRepo.UpdateBuildpackReturns.Error = errors.New("SAD TROMBONE")

			runCommand("my-buildpack1", "new-buildpack")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming buildpack", "my-buildpack"},
				[]string{"SAD TROMBONE"},
			))
		})
	})
})
