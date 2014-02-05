package space_test

import (
	"cf"
	. "cf/commands/space"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
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

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space2 := cf.SpaceFields{}
	space2.Name = "my-space"

	org := cf.OrganizationFields{}
	org.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        space2,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewRenameSpace(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func init() {
	Describe("Testing with ginkgo", func() {

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
			assert.Equal(mr.T(), reqFactory.SpaceName, "my-space")
		})

		It("TestRenameSpaceRun", func() {
			spaceRepo := &testapi.FakeSpaceRepository{}
			space := cf.Space{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: space}
			ui := callRenameSpace(mr.T(), []string{"my-space", "my-new-space"}, reqFactory, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Renaming space", "my-space", "my-new-space", "my-org", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), spaceRepo.RenameSpaceGuid, "my-space-guid")
			assert.Equal(mr.T(), spaceRepo.RenameNewName, "my-new-space")
		})
	})
}
