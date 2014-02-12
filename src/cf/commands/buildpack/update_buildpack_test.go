package buildpack_test

import (
	. "cf/commands/buildpack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callUpdateBuildpack(args []string, reqFactory *testreq.FakeReqFactory, fakeRepo *testapi.FakeBuildpackRepository,
	fakeBitsRepo *testapi.FakeBuildpackBitsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("update-buildpack", args)

	cmd := NewUpdateBuildpack(ui, fakeRepo, fakeBitsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUpdateBuildpackRequirements", func() {
		repo, bitsRepo := getRepositories()

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}
		callUpdateBuildpack([]string{"my-buildpack", "-p", "buildpack.zip", "extraArg"}, reqFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: false}
		callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, BuildpackSuccess: true}
		callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestUpdateBuildpack", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Updating buildpack", "my-buildpack"},
			{"OK"},
		})
	})
	It("TestUpdateBuildpackPosition", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"-i", "999", "my-buildpack"}, reqFactory, repo, bitsRepo)

		Expect(*repo.UpdateBuildpack.Position).To(Equal(999))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Updating buildpack", "my-buildpack"},
			{"OK"},
		})
	})
	It("TestUpdateBuildpackNoPosition", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)

		assert.Nil(mr.T(), repo.UpdateBuildpack.Position)
	})
	It("TestUpdateBuildpackEnabled", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		fakeUI := callUpdateBuildpack([]string{"--enable", "my-buildpack"}, reqFactory, repo, bitsRepo)

		assert.NotNil(mr.T(), repo.UpdateBuildpack.Enabled)
		Expect(*repo.UpdateBuildpack.Enabled).To(Equal(true))

		assert.Contains(mr.T(), fakeUI.Outputs[0], "Updating buildpack")
		assert.Contains(mr.T(), fakeUI.Outputs[0], "my-buildpack")
		assert.Contains(mr.T(), fakeUI.Outputs[1], "OK")
	})
	It("TestUpdateBuildpackDisabled", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		callUpdateBuildpack([]string{"--disable", "my-buildpack"}, reqFactory, repo, bitsRepo)

		assert.NotNil(mr.T(), repo.UpdateBuildpack.Enabled)
		Expect(*repo.UpdateBuildpack.Enabled).To(Equal(false))
	})
	It("TestUpdateBuildpackNoEnable", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		callUpdateBuildpack([]string{"my-buildpack"}, reqFactory, repo, bitsRepo)

		assert.Nil(mr.T(), repo.UpdateBuildpack.Enabled)
	})
	It("TestUpdateBuildpackPath", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"-p", "buildpack.zip", "my-buildpack"}, reqFactory, repo, bitsRepo)

		Expect(bitsRepo.UploadBuildpackPath).To(Equal("buildpack.zip"))

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Updating buildpack", "my-buildpack"},
			{"OK"},
		})
	})
	It("TestUpdateBuildpackWithInvalidPath", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()
		bitsRepo.UploadBuildpackErr = true

		ui := callUpdateBuildpack([]string{"-p", "bogus/path", "my-buildpack"}, reqFactory, repo, bitsRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Updating buildpack", "my-buildpack"},
			{"FAILED"},
		})
	})
	It("TestUpdateBuildpackLock", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"--lock", "my-buildpack"}, reqFactory, repo, bitsRepo)

		assert.NotNil(mr.T(), repo.UpdateBuildpack.Locked)
		Expect(*repo.UpdateBuildpack.Locked).To(Equal(true))

		assert.Contains(mr.T(), ui.Outputs[0], "Updating buildpack")
		assert.Contains(mr.T(), ui.Outputs[0], "my-buildpack")
		assert.Contains(mr.T(), ui.Outputs[1], "OK")
	})
	It("TestUpdateBuildpackUnlock", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"--unlock", "my-buildpack"}, reqFactory, repo, bitsRepo)

		assert.Contains(mr.T(), ui.Outputs[0], "Updating buildpack")
		assert.Contains(mr.T(), ui.Outputs[0], "my-buildpack")
		assert.Contains(mr.T(), ui.Outputs[1], "OK")
	})
	It("TestUpdateBuildpackInvalidLockWithBits", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"--lock", "-p", "buildpack.zip", "my-buildpack"}, reqFactory, repo, bitsRepo)

		assert.Contains(mr.T(), ui.Outputs[1], "FAILED")
	})
	It("TestUpdateBuildpackInvalidUnlockWithBits", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo, bitsRepo := getRepositories()

		ui := callUpdateBuildpack([]string{"--unlock", "-p", "buildpack.zip", "my-buildpack"}, reqFactory, repo, bitsRepo)

		assert.Contains(mr.T(), ui.Outputs[1], "FAILED")
	})
})
