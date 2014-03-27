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

func callRenameBuildpack(args []string, reqFactory *testreq.FakeReqFactory, fakeRepo *testapi.FakeBuildpackRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename-buildpack", args)

	cmd := NewRenameBuildpack(ui, fakeRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestRenameBuildpackRequirements", func() {
		repo := &testapi.FakeBuildpackRepository{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		callRenameBuildpack([]string{"my-buildpack", "my-renamed-buildpack"}, reqFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestRenameBuildpack", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, BuildpackSuccess: true}
		repo := &testapi.FakeBuildpackRepository{}

		ui := callRenameBuildpack([]string{"my-buildpack", "my-renamed-buildpack"}, reqFactory, repo)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Renaming buildpack", "my-buildpack"},
			{"OK"},
		})
	})

})
