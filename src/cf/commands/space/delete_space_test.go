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

func defaultDeleteSpaceSpace() cf.Space {
	space := cf.Space{}
	space.Name = "space-to-delete"
	space.Guid = "space-to-delete-guid"
	return space
}

func defaultDeleteSpaceReqFactory() *testreq.FakeReqFactory {
	return &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: defaultDeleteSpaceSpace()}
}

func deleteSpace(t mr.TestingT, inputs []string, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, spaceRepo *testapi.FakeSpaceRepository) {
	spaceRepo = &testapi.FakeSpaceRepository{}

	ui = &testterm.FakeUI{
		Inputs: inputs,
	}
	ctxt := testcmd.NewContext("delete-space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewDeleteSpace(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func init() {
	Describe("Testing with ginkgo", func() {

		It("TestDeleteSpaceRequirements", func() {
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
			deleteSpace(mr.T(), []string{"y"}, []string{"my-space"}, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
			deleteSpace(mr.T(), []string{"y"}, []string{"my-space"}, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			deleteSpace(mr.T(), []string{"y"}, []string{"my-space"}, reqFactory)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.SpaceName, "my-space")
		})

		It("TestDeleteSpaceConfirmingWithY", func() {

			ui, spaceRepo := deleteSpace(mr.T(), []string{"y"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete", "space-to-delete"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting space", "space-to-delete", "my-org", "my-user"},
				{"OK"},
			})
			assert.Equal(mr.T(), spaceRepo.DeletedSpaceGuid, "space-to-delete-guid")
		})

		It("TestDeleteSpaceConfirmingWithYes", func() {

			ui, spaceRepo := deleteSpace(mr.T(), []string{"Yes"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete", "space-to-delete"},
			})
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting space", "space-to-delete", "my-org", "my-user"},
				{"OK"},
			})
			assert.Equal(mr.T(), spaceRepo.DeletedSpaceGuid, "space-to-delete-guid")
		})

		It("TestDeleteSpaceWithForceOption", func() {

			ui, spaceRepo := deleteSpace(mr.T(), []string{}, []string{"-f", "space-to-delete"}, defaultDeleteSpaceReqFactory())

			assert.Equal(mr.T(), len(ui.Prompts), 0)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "space-to-delete"},
				{"OK"},
			})
			assert.Equal(mr.T(), spaceRepo.DeletedSpaceGuid, "space-to-delete-guid")
		})

		It("TestDeleteSpaceWhenSpaceIsTargeted", func() {

			reqFactory := defaultDeleteSpaceReqFactory()
			spaceRepo := &testapi.FakeSpaceRepository{}

			config := &configuration.Configuration{}
			config.SpaceFields = defaultDeleteSpaceSpace().SpaceFields

			ui := &testterm.FakeUI{}
			ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

			cmd := NewDeleteSpace(ui, config, spaceRepo)
			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.Equal(mr.T(), config.HasSpace(), false)
		})

		It("TestDeleteSpaceWhenSpaceNotTargeted", func() {
			reqFactory := defaultDeleteSpaceReqFactory()
			spaceRepo := &testapi.FakeSpaceRepository{}

			otherSpace := cf.SpaceFields{}
			otherSpace.Name = "do-not-delete"
			otherSpace.Guid = "do-not-delete-guid"

			config := &configuration.Configuration{}
			config.SpaceFields = otherSpace

			ui := &testterm.FakeUI{}
			ctxt := testcmd.NewContext("delete", []string{"-f", "space-to-delete"})

			cmd := NewDeleteSpace(ui, config, spaceRepo)
			testcmd.RunCommand(cmd, ctxt, reqFactory)

			assert.Equal(mr.T(), config.HasSpace(), true)
		})

		It("TestDeleteSpaceCommandWith", func() {

			ui, _ := deleteSpace(mr.T(), []string{"Yes"}, []string{}, defaultDeleteSpaceReqFactory())
			assert.True(mr.T(), ui.FailedWithUsage)

			ui, _ = deleteSpace(mr.T(), []string{"Yes"}, []string{"space-to-delete"}, defaultDeleteSpaceReqFactory())
			assert.False(mr.T(), ui.FailedWithUsage)
		})
	})
}
