package buildpack_test

import (
	. "cf/commands/buildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("create-buildpack command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		repo                *testapi.FakeBuildpackRepository
		bitsRepo            *testapi.FakeBuildpackBitsRepository
		ui                  *testterm.FakeUI
		cmd                 CreateBuildpack
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		repo = &testapi.FakeBuildpackRepository{}
		bitsRepo = &testapi.FakeBuildpackBitsRepository{}
		ui = &testterm.FakeUI{}
		cmd = NewCreateBuildpack(ui, repo, bitsRepo)
	})

	It("fails requirements when the user is not logged in", func() {
		requirementsFactory.LoginSuccess = false
		context := testcmd.NewContext("create-buildpack", []string{"my-buildpack", "my-dir", "0"})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("fails with usage when given fewer than three arguments", func() {
		context := testcmd.NewContext("create-buildpack", []string{})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("creates and uploads buildpacks", func() {
		context := testcmd.NewContext("create-buildpack", []string{"my-buildpack", "my.war", "5"})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		Expect(repo.CreateBuildpack.Enabled).To(BeNil())
		Expect(ui.FailedWithUsage).To(BeFalse())

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating buildpack", "my-buildpack"},
			{"OK"},
			{"Uploading buildpack", "my-buildpack"},
			{"OK"},
		})
		testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
			{"FAILED"},
		})
	})

	It("warns the user when the buildpack already exists", func() {
		repo.CreateBuildpackExists = true
		context := testcmd.NewContext("create-buildpack", []string{"my-buildpack", "my.war", "5"})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating buildpack", "my-buildpack"},
			{"OK"},
			{"my-buildpack", "already exists"},
			{"tip", "update-buildpack"},
		})
		testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
			{"FAILED"},
		})
	})

	It("enables the buildpack when given the --enabled flag", func() {
		context := testcmd.NewContext("create-buildpack", []string{"--enable", "my-buildpack", "my.war", "5"})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		Expect(*repo.CreateBuildpack.Enabled).To(Equal(true))
	})

	It("disables the buildpack when given the --disable flag", func() {
		context := testcmd.NewContext("create-buildpack", []string{"--disable", "my-buildpack", "my.war", "5"})
		testcmd.RunCommand(cmd, context, requirementsFactory)
		Expect(*repo.CreateBuildpack.Enabled).To(Equal(false))
	})

	It("alerts the user when uploading the buildpack bits fails", func() {
		bitsRepo.UploadBuildpackErr = true
		context := testcmd.NewContext("create-buildpack", []string{"my-buildpack", "bogus/path", "5"})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating buildpack", "my-buildpack"},
			{"OK"},
			{"Uploading buildpack"},
			{"FAILED"},
		})
	})
})
