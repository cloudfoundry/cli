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
	space := cf.Space{}
	space.Name = "space1"
	space2 := cf.Space{}
	space2.Name = "space2"
	space3 := cf.Space{}
	space3.Name = "space3"
	spaceRepo := &testapi.FakeSpaceRepository{
		Spaces: []cf.Space{space, space2, space3},
	}
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})

	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		OrganizationFields: org,
		AccessToken:        token,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	ui := callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.Contains(t, ui.Outputs[0], "Getting spaces in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[2], "space1")
	assert.Contains(t, ui.Outputs[3], "space2")
	assert.Contains(t, ui.Outputs[4], "space3")
}

func TestListingSpacesWhenNoSpaces(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{
		Spaces: []cf.Space{},
	}
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})

	assert.NoError(t, err)
	org2 := cf.OrganizationFields{}
	org2.Name = "my-org"
	config := &configuration.Configuration{
		OrganizationFields: org2,
		AccessToken:        token,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	ui := callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.Contains(t, ui.Outputs[0], "Getting spaces in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "No spaces found")
}

func callSpaces(args []string, reqFactory *testreq.FakeReqFactory, config *configuration.Configuration, spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("spaces", args)

	cmd := NewListSpaces(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
