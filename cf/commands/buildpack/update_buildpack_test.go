package buildpack_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func successfulUpdate(ui *testterm.FakeUI, buildpackName string) {
	Expect(ui.Outputs()).To(ContainSubstrings(
		[]string{"Updating buildpack", buildpackName},
		[]string{"OK"},
	))
}

func failedUpdate(ui *testterm.FakeUI, buildpackName string) {
	Expect(ui.Outputs()).To(ContainSubstrings(
		[]string{"Updating buildpack", buildpackName},
		[]string{"FAILED"},
	))
}

var _ = Describe("Updating buildpack command", func() {
	var (
		requirementsFactory *requirementsfakes.FakeFactory
		ui                  *testterm.FakeUI
		repo                *apifakes.OldFakeBuildpackRepository
		bitsRepo            *apifakes.FakeBuildpackBitsRepository
		deps                commandregistry.Dependency

		buildpackName string
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetBuildpackRepository(repo)
		deps.RepoLocator = deps.RepoLocator.SetBuildpackBitsRepository(bitsRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("update-buildpack").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		buildpackName = "my-buildpack"

		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		buildpackReq := new(requirementsfakes.FakeBuildpackRequirement)
		buildpackReq.GetBuildpackReturns(models.Buildpack{Name: buildpackName})
		requirementsFactory.NewBuildpackRequirementReturns(buildpackReq)
		ui = new(testterm.FakeUI)
		repo = new(apifakes.OldFakeBuildpackRepository)
		bitsRepo = new(apifakes.FakeBuildpackBitsRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("update-buildpack", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Context("is only successful on login and buildpack success", func() {
		It("returns success when both are true", func() {
			Expect(runCommand(buildpackName)).To(BeTrue())
		})

		It("returns failure when at requirements error", func() {
			buildpackReq := new(requirementsfakes.FakeBuildpackRequirement)
			buildpackReq.ExecuteReturns(errors.New("no build pack"))
			requirementsFactory.NewBuildpackRequirementReturns(buildpackReq)

			Expect(runCommand(buildpackName, "-p", "buildpack.zip", "extraArg")).To(BeFalse())
		})
	})

	Context("when a file is provided", func() {
		It("prints error and do not call create buildpack", func() {
			bitsRepo.CreateBuildpackZipFileReturns(nil, "", fmt.Errorf("create buildpack error"))

			Expect(runCommand(buildpackName, "-p", "file")).To(BeFalse())

			Expect(ui.Outputs()).To(ContainSubstrings([]string{"Failed to create a local temporary zip file for the buildpack"}))
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
			Expect(bitsRepo.UploadBuildpackCallCount()).To(Equal(0))

		})
	})

	Context("when a path is provided", func() {
		It("updates buildpack", func() {
			runCommand(buildpackName)

			successfulUpdate(ui, buildpackName)
		})
	})

	Context("when a URL is provided", func() {
		It("updates buildpack", func() {
			testcmd.RunCLICommand("update-buildpack", []string{"my-buildpack", "-p", "https://some-url.com"}, requirementsFactory, updateCommandDependency, false, ui)

			Expect(bitsRepo.CreateBuildpackZipFileCallCount()).To(Equal(1))
			buildpackPath := bitsRepo.CreateBuildpackZipFileArgsForCall(0)
			Expect(buildpackPath).To(Equal("https://some-url.com"))
			successfulUpdate(ui, buildpackName)
		})
	})

	Context("updates buildpack when passed the proper flags", func() {
		Context("position flag", func() {
			It("sets the position when passed a value", func() {
				runCommand("-i", "999", buildpackName)

				Expect(*repo.UpdateBuildpackArgs.Buildpack.Position).To(Equal(999))
				successfulUpdate(ui, buildpackName)
			})

			It("defaults to nil when not passed", func() {
				runCommand(buildpackName)

				Expect(repo.UpdateBuildpackArgs.Buildpack.Position).To(BeNil())
			})
		})

		Context("enabling/disabling buildpacks", func() {
			It("can enable buildpack", func() {
				runCommand("--enable", buildpackName)

				Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).NotTo(BeNil())
				Expect(*repo.UpdateBuildpackArgs.Buildpack.Enabled).To(Equal(true))

				successfulUpdate(ui, buildpackName)
			})

			It("can disable buildpack", func() {
				runCommand("--disable", buildpackName)

				Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).NotTo(BeNil())
				Expect(*repo.UpdateBuildpackArgs.Buildpack.Enabled).To(Equal(false))
			})

			It("defaults to nil when not passed", func() {
				runCommand(buildpackName)

				Expect(repo.UpdateBuildpackArgs.Buildpack.Enabled).To(BeNil())
			})
		})

		Context("buildpack path", func() {
			It("uploads buildpack when passed", func() {
				runCommand("-p", "buildpack.zip", buildpackName)
				Expect(bitsRepo.CreateBuildpackZipFileCallCount()).To(Equal(1))
				buildpackPath := bitsRepo.CreateBuildpackZipFileArgsForCall(0)
				Expect(buildpackPath).To(Equal("buildpack.zip"))

				successfulUpdate(ui, buildpackName)
			})

			It("errors when passed invalid path", func() {
				bitsRepo.UploadBuildpackReturns(fmt.Errorf("upload error"))

				runCommand("-p", "bogus/path", buildpackName)

				failedUpdate(ui, buildpackName)
			})
		})

		Context("locking buildpack", func() {
			It("can lock a buildpack", func() {
				runCommand("--lock", buildpackName)

				Expect(repo.UpdateBuildpackArgs.Buildpack.Locked).NotTo(BeNil())
				Expect(*repo.UpdateBuildpackArgs.Buildpack.Locked).To(Equal(true))

				successfulUpdate(ui, buildpackName)
			})

			It("can unlock a buildpack", func() {
				runCommand("--unlock", buildpackName)

				successfulUpdate(ui, buildpackName)
			})

			Context("Unsuccessful locking", func() {
				It("lock fails when passed invalid path", func() {
					runCommand("--lock", "-p", "buildpack.zip", buildpackName)

					failedUpdate(ui, buildpackName)
				})

				It("unlock fails when passed invalid path", func() {
					runCommand("--unlock", "-p", "buildpack.zip", buildpackName)

					failedUpdate(ui, buildpackName)
				})
			})
		})
	})
})
