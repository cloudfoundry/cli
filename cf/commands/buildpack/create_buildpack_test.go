package buildpack_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-buildpack command", func() {
	var (
		requirementsFactory *requirementsfakes.FakeFactory
		repo                *apifakes.OldFakeBuildpackRepository
		bitsRepo            *apifakes.FakeBuildpackBitsRepository
		ui                  *testterm.FakeUI
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetBuildpackRepository(repo)
		deps.RepoLocator = deps.RepoLocator.SetBuildpackBitsRepository(bitsRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("create-buildpack").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		repo = new(apifakes.OldFakeBuildpackRepository)
		bitsRepo = new(apifakes.FakeBuildpackBitsRepository)
		ui = &testterm.FakeUI{}
	})

	It("fails requirements when the user is not logged in", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		Expect(testcmd.RunCLICommand("create-buildpack", []string{"my-buildpack", "my-dir", "0"}, requirementsFactory, updateCommandDependency, false, ui)).To(BeFalse())
	})

	It("fails with usage when given fewer than three arguments", func() {
		testcmd.RunCLICommand("create-buildpack", []string{}, requirementsFactory, updateCommandDependency, false, ui)
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
	})

	Context("when a file is provided", func() {
		It("prints error and do not call create buildpack", func() {
			bitsRepo.CreateBuildpackZipFileReturns(nil, "", fmt.Errorf("create buildpack error"))

			testcmd.RunCLICommand("create-buildpack", []string{"my-buildpack", "file", "5"}, requirementsFactory, updateCommandDependency, false, ui)

			Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"Failed to create a local temporary zip file for the buildpack"}))
			Expect(ui.Outputs()).NotTo(ContainSubstrings([]string{"Creating buildpack"}))

		})
	})

	Context("when a directory is provided", func() {
		It("creates and uploads buildpacks", func() {
			testcmd.RunCLICommand("create-buildpack", []string{"my-buildpack", "my.war", "5"}, requirementsFactory, updateCommandDependency, false, ui)

			Expect(repo.CreateBuildpack.Enabled).To(BeNil())
			Expect(bitsRepo.CreateBuildpackZipFileCallCount()).To(Equal(1))
			buildpackPath := bitsRepo.CreateBuildpackZipFileArgsForCall(0)
			Expect(buildpackPath).To(Equal("my.war"))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating buildpack", "my-buildpack"},
				[]string{"OK"},
				[]string{"Uploading buildpack", "my-buildpack"},
				[]string{"OK"},
			))
			Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
		})
	})

	Context("when a URL is provided", func() {
		It("creates and uploads buildpacks", func() {
			testcmd.RunCLICommand("create-buildpack", []string{"my-buildpack", "https://some-url.com", "5"}, requirementsFactory, updateCommandDependency, false, ui)

			Expect(repo.CreateBuildpack.Enabled).To(BeNil())
			Expect(bitsRepo.CreateBuildpackZipFileCallCount()).To(Equal(1))
			buildpackPath := bitsRepo.CreateBuildpackZipFileArgsForCall(0)
			Expect(buildpackPath).To(Equal("https://some-url.com"))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating buildpack", "my-buildpack"},
				[]string{"OK"},
				[]string{"Uploading buildpack", "my-buildpack"},
				[]string{"OK"},
			))
			Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
		})
	})

	It("warns the user when the buildpack already exists", func() {
		repo.CreateBuildpackExists = true
		testcmd.RunCLICommand("create-buildpack", []string{"my-buildpack", "my.war", "5"}, requirementsFactory, updateCommandDependency, false, ui)

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Creating buildpack", "my-buildpack"},
			[]string{"OK"},
			[]string{"my-buildpack", "already exists"},
			[]string{"TIP", "use", cf.Name, "update-buildpack"},
		))
		Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"FAILED"}))
	})

	It("enables the buildpack when given the --enabled flag", func() {
		testcmd.RunCLICommand("create-buildpack", []string{"--enable", "my-buildpack", "my.war", "5"}, requirementsFactory, updateCommandDependency, false, ui)

		Expect(*repo.CreateBuildpack.Enabled).To(Equal(true))
	})

	It("disables the buildpack when given the --disable flag", func() {
		testcmd.RunCLICommand("create-buildpack", []string{"--disable", "my-buildpack", "my.war", "5"}, requirementsFactory, updateCommandDependency, false, ui)

		Expect(*repo.CreateBuildpack.Enabled).To(Equal(false))
	})

	It("alerts the user when uploading the buildpack bits fails", func() {
		bitsRepo.UploadBuildpackReturns(fmt.Errorf("upload error"))

		testcmd.RunCLICommand("create-buildpack", []string{"my-buildpack", "bogus/path", "5"}, requirementsFactory, updateCommandDependency, false, ui)

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Creating buildpack", "my-buildpack"},
			[]string{"OK"},
			[]string{"Uploading buildpack"},
			[]string{"FAILED"},
		))
	})
})
