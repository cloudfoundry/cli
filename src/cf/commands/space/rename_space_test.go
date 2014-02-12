package space_test

import (
	. "cf/commands/space"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callRenameSpace(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, spaceRepo *testapi.FakeSpaceRepository) (ui *testterm.FakeUI) {
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

		ui := callRenameSpace(mr.T(), []string{}, reqFactory, spaceRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callRenameSpace(mr.T(), []string{"foo"}, reqFactory, spaceRepo)
		assert.True(mr.T(), ui.FailedWithUsage)
	})

	It("TestRenameSpaceRequirements", func() {

		spaceRepo := &testapi.FakeSpaceRepository{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callRenameSpace(mr.T(), []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callRenameSpace(mr.T(), []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callRenameSpace(mr.T(), []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		Expect(reqFactory.SpaceName).To(Equal("my-space"))
	})

	It("TestRenameSpaceRun", func() {
		spaceRepo := &testapi.FakeSpaceRepository{}
		space := models.Space{}
		space.Name = "my-space"
		space.Guid = "my-space-guid"
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: space}
		ui := callRenameSpace(mr.T(), []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Renaming space", "my-space", "my-new-space", "my-org", "my-user"},
			{"OK"},
		})

		Expect(spaceRepo.RenameSpaceGuid).To(Equal("my-space-guid"))
		Expect(spaceRepo.RenameNewName).To(Equal("my-new-space"))
	})
})
