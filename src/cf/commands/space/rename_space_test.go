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
	"cf/configuration"
)


var _ = Describe("rename-space command", func() {
	var (
		ui *testterm.FakeUI
		configRepo configuration.ReadWriter
		reqFactory *testreq.FakeReqFactory
		spaceRepo *testapi.FakeSpaceRepository
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		spaceRepo = &testapi.FakeSpaceRepository{}
	})

	var callRenameSpace = func(args []string) {
		cmd := NewRenameSpace(ui, configRepo, spaceRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("create-space", args), reqFactory)
	}

	Describe("when the user is not logged in", func() {
		It("does not pass requirements", func() {
			reqFactory.LoginSuccess = false
			callRenameSpace([]string{"my-space", "my-new-space"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Describe("when the user has not targeted an org", func() {
		It("does not pass requirements", func() {
			reqFactory.TargetedOrgSuccess = false
			callRenameSpace([]string{"my-space", "my-new-space"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Describe("when the user provides fewer than two args", func() {
		It("fails with usage", func() {
			callRenameSpace([]string{"foo"})
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	It("renames a space", func() {
		space := models.Space{}
		space.Name = "my-space"
		space.Guid = "my-space-guid"
		reqFactory.Space = space

		callRenameSpace([]string{"my-space", "my-new-space"})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Renaming space", "my-space", "my-new-space", "my-org", "my-user"},
			{"OK"},
		})

		Expect(spaceRepo.RenameSpaceGuid).To(Equal("my-space-guid"))
		Expect(spaceRepo.RenameNewName).To(Equal("my-new-space"))
	})
})
