package buildpack_test

import (
	"strings"

	"github.com/cloudfoundry/cli/cf"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-buildpack command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		repo                *testapi.FakeBuildpackRepository
		bitsRepo            *testapi.FakeBuildpackBitsRepository
		ui                  *testterm.FakeUI
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetBuildpackRepository(repo)
		deps.RepoLocator = deps.RepoLocator.SetBuildpackBitsRepository(bitsRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-buildpack").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		repo = &testapi.FakeBuildpackRepository{}
		bitsRepo = &testapi.FakeBuildpackBitsRepository{}
		ui = &testterm.FakeUI{}
	})

	It("fails requirements when the user is not logged in", func() {
		requirementsFactory.LoginSuccess = false
		Expect(testcmd.RunCliCommand("create-buildpack", []string{"my-buildpack", "my-dir", "0"}, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
	})

	It("fails with usage when given fewer than three arguments", func() {
		testcmd.RunCliCommand("create-buildpack", []string{}, requirementsFactory, updateCommandDependency, false)
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
	})

	It("creates and uploads buildpacks", func() {
		testcmd.RunCliCommand("create-buildpack", []string{"my-buildpack", "my.war", "5"}, requirementsFactory, updateCommandDependency, false)

		Expect(repo.CreateBuildpack.Enabled).To(BeNil())
		Expect(strings.HasSuffix(bitsRepo.UploadBuildpackPath, "my.war")).To(Equal(true))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating buildpack", "my-buildpack"},
			[]string{"OK"},
			[]string{"Uploading buildpack", "my-buildpack"},
			[]string{"OK"},
		))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
	})

	It("warns the user when the buildpack already exists", func() {
		repo.CreateBuildpackExists = true
		testcmd.RunCliCommand("create-buildpack", []string{"my-buildpack", "my.war", "5"}, requirementsFactory, updateCommandDependency, false)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating buildpack", "my-buildpack"},
			[]string{"OK"},
			[]string{"my-buildpack", "already exists"},
			[]string{"TIP", "use", cf.Name(), "update-buildpack"},
		))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"FAILED"}))
	})

	It("enables the buildpack when given the --enabled flag", func() {
		testcmd.RunCliCommand("create-buildpack", []string{"--enable", "my-buildpack", "my.war", "5"}, requirementsFactory, updateCommandDependency, false)

		Expect(*repo.CreateBuildpack.Enabled).To(Equal(true))
	})

	It("disables the buildpack when given the --disable flag", func() {
		testcmd.RunCliCommand("create-buildpack", []string{"--disable", "my-buildpack", "my.war", "5"}, requirementsFactory, updateCommandDependency, false)

		Expect(*repo.CreateBuildpack.Enabled).To(Equal(false))
	})

	It("alerts the user when uploading the buildpack bits fails", func() {
		bitsRepo.UploadBuildpackErr = true
		testcmd.RunCliCommand("create-buildpack", []string{"my-buildpack", "bogus/path", "5"}, requirementsFactory, updateCommandDependency, false)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating buildpack", "my-buildpack"},
			[]string{"OK"},
			[]string{"Uploading buildpack"},
			[]string{"FAILED"},
		))
	})
})
