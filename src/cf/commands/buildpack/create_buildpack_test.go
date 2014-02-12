package buildpack_test

import (
	. "cf/commands/buildpack"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func getRepositories() (*testapi.FakeBuildpackRepository, *testapi.FakeBuildpackBitsRepository) {
	return &testapi.FakeBuildpackRepository{}, &testapi.FakeBuildpackBitsRepository{}
}

func callCreateBuildpack(args []string, reqFactory *testreq.FakeReqFactory, fakeRepo *testapi.FakeBuildpackRepository,
	fakeBitsRepo *testapi.FakeBuildpackBitsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-buildpack", args)

	cmd := NewCreateBuildpack(ui, fakeRepo, fakeBitsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateBuildpackRequirements", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()

		repo.FindByNameBuildpack = models.Buildpack{}
		callCreateBuildpack([]string{"my-buildpack", "my-dir", "0"}, reqFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory.LoginSuccess = false
		callCreateBuildpack([]string{"my-buildpack", "my-dir", "0"}, reqFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestCreateBuildpack", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()
		ui := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

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
	It("TestCreateBuildpackWhenItAlreadyExists", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()

		repo.CreateBuildpackExists = true
		ui := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

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
	It("TestCreateBuildpackWithPosition", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()
		ui := callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

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
	It("TestCreateBuildpackEnabled", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()
		ui := callCreateBuildpack([]string{"--enable", "my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

		Expect(repo.CreateBuildpack.Enabled).NotTo(BeNil())
		Expect(*repo.CreateBuildpack.Enabled).To(Equal(true))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"creating buildpack", "my-buildpack"},
			{"OK"},
			{"uploading buildpack", "my-buildpack"},
			{"OK"},
		})
		testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
			{"FAILED"},
		})
	})
	It("TestCreateBuildpackNoEnableFlag", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()
		callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

		Expect(repo.CreateBuildpack.Enabled).To(BeNil())
	})
	It("TestCreateBuildpackDisabled", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()
		callCreateBuildpack([]string{"--disable", "my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)

		Expect(repo.CreateBuildpack.Enabled).NotTo(BeNil())
		Expect(*repo.CreateBuildpack.Enabled).To(Equal(false))
	})
	It("TestCreateBuildpackWithInvalidPath", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()

		bitsRepo.UploadBuildpackErr = true
		ui := callCreateBuildpack([]string{"my-buildpack", "bogus/path", "5"}, reqFactory, repo, bitsRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating buildpack", "my-buildpack"},
			{"OK"},
			{"Uploading buildpack"},
			{"FAILED"},
		})
	})
	It("TestCreateBuildpackFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callCreateBuildpack([]string{}, reqFactory, repo, bitsRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateBuildpack([]string{"my-buildpack", "my.war", "5"}, reqFactory, repo, bitsRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
})
