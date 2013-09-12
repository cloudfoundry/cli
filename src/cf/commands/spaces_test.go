package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestSpacesRequirements(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, OrgSuccess: true}
	callSpaces([]string{}, reqFactory, spaceRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, OrgSuccess: false}
	callSpaces([]string{}, reqFactory, spaceRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, OrgSuccess: true}
	callSpaces([]string{}, reqFactory, spaceRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestListingSpaces(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{
		Spaces: []cf.Space{
			cf.Space{Name: "space1"}, cf.Space{Name: "space2"},
		},
	}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, OrgSuccess: true}
	ui := callSpaces([]string{}, reqFactory, spaceRepo)
	assert.Contains(t, ui.Outputs[0], "Getting spaces...")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "space1")
	assert.Contains(t, ui.Outputs[3], "space2")
}

func callSpaces(args []string, reqFactory *testhelpers.FakeReqFactory, spaceRepo api.SpaceRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("spaces", args)

	cmd := NewSpaces(ui, spaceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
