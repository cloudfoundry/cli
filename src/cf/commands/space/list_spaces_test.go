package space_test

import (
	"cf"
	"cf/api"
	. "cf/commands/space"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSpacesRequirements(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}
	config := &configuration.Configuration{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestListingSpaces(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{
		Spaces: []cf.Space{
			cf.Space{Name: "space1"}, cf.Space{Name: "space2"},
		},
	}
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})

	assert.NoError(t, err)
	config := &configuration.Configuration{
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	ui := callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.Contains(t, ui.Outputs[0], "Getting spaces in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "space1")
	assert.Contains(t, ui.Outputs[3], "space2")
}

func callSpaces(args []string, reqFactory *testreq.FakeReqFactory, config *configuration.Configuration, spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("spaces", args)

	cmd := NewListSpaces(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
