package space_test

import (
	. "cf/commands/space"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callRenameSpace(args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-space", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewRenameSpace(ui, configRepo, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {

	It("TestRenameSpaceFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}
		spaceRepo := &testapi.FakeSpaceRepository{}

		ui := callRenameSpace([]string{}, reqFactory, spaceRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callRenameSpace([]string{"foo"}, reqFactory, spaceRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("TestRenameSpaceRequirements", func() {

		spaceRepo := &testapi.FakeSpaceRepository{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callRenameSpace([]string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callRenameSpace([]string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callRenameSpace([]string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.SpaceName).To(Equal("my-space"))
	})

	It("TestRenameSpaceRun", func() {
		spaceRepo := &testapi.FakeSpaceRepository{}
		space := models.Space{}
		space.Name = "my-space"
		space.Guid = "my-space-guid"
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: space}
		ui := callRenameSpace([]string{"my-space", "my-new-space"}, reqFactory, spaceRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Renaming space", "my-space", "my-new-space", "my-org", "my-user"},
			{"OK"},
		})

		Expect(spaceRepo.RenameSpaceGuid).To(Equal("my-space-guid"))
		Expect(spaceRepo.RenameNewName).To(Equal("my-new-space"))
	})
})
