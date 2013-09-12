package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestSpacesRequirements(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	config := &configuration.Configuration{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, OrgSuccess: true}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, OrgSuccess: false}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, OrgSuccess: true}
	callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestListingSpaces(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{
		Spaces: []cf.Space{
			cf.Space{Name: "space1"}, cf.Space{Name: "space2"},
		},
	}
	config := &configuration.Configuration{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, OrgSuccess: true}
	ui := callSpaces([]string{}, reqFactory, config, spaceRepo)
	assert.Contains(t, ui.Outputs[0], "Getting spaces")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "space1")
	assert.Contains(t, ui.Outputs[3], "space2")
}

func callSpaces(args []string, reqFactory *testhelpers.FakeReqFactory, config *configuration.Configuration, spaceRepo api.SpaceRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("spaces", args)

	cmd := NewSpaces(ui, config, spaceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
