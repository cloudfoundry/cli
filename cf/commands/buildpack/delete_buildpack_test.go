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

var _ = Describe("delete-buildpack command", func() {
	var (
		ui                  *testterm.FakeUI
		buildpackRepo       *testapi.FakeBuildpackRepository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetBuildpackRepository(buildpackRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("delete-buildpack").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		buildpackRepo = &testapi.FakeBuildpackRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("delete-buildpack", args, requirementsFactory, updateCommandDependency, false)
	}

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = false
		})

		It("fails requirements", func() {
			Expect(runCommand("-f", "my-buildpack")).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when the buildpack exists", func() {
			BeforeEach(func() {
				buildpackRepo.FindByNameBuildpack = models.Buildpack{
					Name: "my-buildpack",
					Guid: "my-buildpack-guid",
				}
			})

			It("deletes the buildpack", func() {
				ui = &testterm.FakeUI{Inputs: []string{"y"}}

				runCommand("my-buildpack")

				Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

				Expect(ui.Prompts).To(ContainSubstrings([]string{"delete the buildpack my-buildpack"}))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting buildpack", "my-buildpack"},
					[]string{"OK"},
				))
			})

			Context("when the force flag is provided", func() {
				It("does not prompt the user to delete the buildback", func() {
					runCommand("-f", "my-buildpack")

					Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

					Expect(len(ui.Prompts)).To(Equal(0))
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Deleting buildpack", "my-buildpack"},
						[]string{"OK"},
					))
				})
			})
		})

		Context("when the buildpack provided is not found", func() {
			BeforeEach(func() {
				ui = &testterm.FakeUI{Inputs: []string{"y"}}
				buildpackRepo.FindByNameNotFound = true
			})

			It("warns the user", func() {
				runCommand("my-buildpack")

				Expect(buildpackRepo.FindByNameName).To(Equal("my-buildpack"))
				Expect(buildpackRepo.FindByNameNotFound).To(BeTrue())

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "my-buildpack"},
					[]string{"OK"},
				))

				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"my-buildpack", "does not exist"}))
			})
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				ui = &testterm.FakeUI{Inputs: []string{"y"}}

				buildpackRepo.FindByNameBuildpack = models.Buildpack{
					Name: "my-buildpack",
					Guid: "my-buildpack-guid",
				}
				buildpackRepo.DeleteApiResponse = errors.New("failed badly")
			})

			It("fails with the error", func() {
				runCommand("my-buildpack")

				Expect(buildpackRepo.DeleteBuildpackGuid).To(Equal("my-buildpack-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting buildpack", "my-buildpack"},
					[]string{"FAILED"},
					[]string{"my-buildpack"},
					[]string{"failed badly"},
				))
			})
		})
	})
})
