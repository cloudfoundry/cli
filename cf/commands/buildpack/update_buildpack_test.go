package buildpack_test

import (
	"strings"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func successfulUpdate(ui *testterm.FakeUI, buildpackName string) {
	Expect(ui.Outputs).To(ContainSubstrings(
		[]string{"Updating buildpack", buildpackName},
		[]string{"OK"},
	))
}

func failedUpdate(ui *testterm.FakeUI, buildpackName string) {
	Expect(ui.Outputs).To(ContainSubstrings(
		[]string{"Updating buildpack", buildpackName},
		[]string{"FAILED"},
	))
}

var _ = Describe("Updating buildpack command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
		repo                *testapi.FakeBuildpackRepository
		bitsRepo            *testapi.FakeBuildpackBitsRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetBuildpackRepository(repo)
		deps.RepoLocator = deps.RepoLocator.SetBuildpackBitsRepository(bitsRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("update-buildpack").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		ui = new(testterm.FakeUI)
		repo = &testapi.FakeBuildpackRepository{}
		bitsRepo = &testapi.FakeBuildpackBitsRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("update-buildpack", args, requirementsFactory, updateCommandDependency, false)
	}

	Context("is only successful on login and buildpack success", func() {
		It("returns success when both are true", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}

			Expect(runCommand("my-buildpack")).To(BeTrue())
		})

		It("returns failure when at least one is false", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}
			Expect(runCommand("my-buildpack", "-p", "buildpack.zip", "extraArg")).To(BeFalse())

			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}

			Expect(runCommand("my-buildpack")).To(BeFalse())

			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false, BuildpackSuccess: true}

			Expect(runCommand("my-buildpack")).To(BeFalse())
		})
	})

	It("updates buildpack", func() {
		runCommand("my-buildpack")

		successfulUpdate(ui, "my-buildpack")
	})

	Context("updates buildpack when passed the proper flags", func() {
		Context("position flag", func() {
			It("sets the position when passed a value", func() {
				runCommand("-i", "999", "my-buildpack")

				Expect(*repo.UpdateBuildpackArgs.Buildpack.Position).To(Equal(999))
				successfulUpdate(ui, "my-buildpack")
			})

			It("defaults to nil when not passed", func() {
				runCommand("my-buildpack")

				Expect(repo.UpdateBuildpackArgs.Buildpack.Position).To(BeNil())
			})
		})

		Context("enabling/disabling buildpacks", func() {
			It("can enable buildpack", func() {
				runCommand("--enable", "my-buildpack")

				Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).NotTo(BeNil())
				Expect(*repo.UpdateBuildpackArgs.Buildpack.Enabled).To(Equal(true))

				successfulUpdate(ui, "my-buildpack")
			})

			It("can disable buildpack", func() {
				runCommand("--disable", "my-buildpack")

				Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).NotTo(BeNil())
				Expect(*repo.UpdateBuildpackArgs.Buildpack.Enabled).To(Equal(false))
			})

			It("defaults to nil when not passed", func() {
				runCommand("my-buildpack")

				Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).To(BeNil())
			})
		})

		Context("buildpack path", func() {
			It("uploads buildpack when passed", func() {
				runCommand("-p", "buildpack.zip", "my-buildpack")
				Expect(strings.HasSuffix(bitsRepo.UploadBuildpackPath, "buildpack.zip")).To(Equal(true))

				successfulUpdate(ui, "my-buildpack")
			})

			It("errors when passed invalid path", func() {
				bitsRepo.UploadBuildpackErr = true

				runCommand("-p", "bogus/path", "my-buildpack")

				failedUpdate(ui, "my-buildpack")
			})
		})

		Context("locking buildpack", func() {
			It("can lock a buildpack", func() {
				runCommand("--lock", "my-buildpack")

				Expect(repo.UpdateBuildpackArgs.Buildpack.Locked).NotTo(BeNil())
				Expect(*repo.UpdateBuildpackArgs.Buildpack.Locked).To(Equal(true))

				successfulUpdate(ui, "my-buildpack")
			})

			It("can unlock a buildpack", func() {
				runCommand("--unlock", "my-buildpack")

				successfulUpdate(ui, "my-buildpack")
			})

			Context("Unsuccessful locking", func() {
				It("lock fails when passed invalid path", func() {
					runCommand("--lock", "-p", "buildpack.zip", "my-buildpack")

					failedUpdate(ui, "my-buildpack")
				})

				It("unlock fails when passed invalid path", func() {
					runCommand("--unlock", "-p", "buildpack.zip", "my-buildpack")

					failedUpdate(ui, "my-buildpack")
				})
			})
		})
	})

})
