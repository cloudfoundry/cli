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

func getTargetDependencies() (orgRepo *testhelpers.FakeOrgRepository, spaceRepo *testhelpers.FakeSpaceRepository, configRepo *testhelpers.FakeConfigRepository, reqFactory *testhelpers.FakeReqFactory) {
	orgRepo = &testhelpers.FakeOrgRepository{}
	spaceRepo = &testhelpers.FakeSpaceRepository{}
	configRepo = &testhelpers.FakeConfigRepository{}
	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true}
	return
}

func TestTargetRequirements(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	reqFactory.LoginSuccess = false

	callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true

	callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
}

func TestTargetWithoutArgumentAndLoggedIn(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	config := configRepo.Login()
	config.Target = "https://api.run.pivotal.io"

	ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, len(ui.Outputs), 4)
	assert.Contains(t, ui.Outputs[0], "https://api.run.pivotal.io")
	assert.Contains(t, ui.Outputs[1], "user: ")
	assert.Contains(t, ui.Outputs[2], "No org targeted")
	assert.Contains(t, ui.Outputs[3], "No space targeted")
}

// Start test with organization option

func TestTargetOrganizationWhenUserHasAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Login()
	config, err := configRepo.Get()
	assert.NoError(t, err)
	config.Space = cf.Space{Name: "my-space", Guid: "my-space-guid"}

	orgs := []cf.Organization{
		cf.Organization{Name: "my-organization", Guid: "my-organization-guid"},
	}
	orgRepo.Organizations = orgs
	orgRepo.FindByNameOrganization = orgs[0]

	ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, orgRepo.FindByNameName, "my-organization")
	assert.Contains(t, ui.Outputs[2], "org:")
	assert.Contains(t, ui.Outputs[2], "my-organization")
	assert.Contains(t, ui.Outputs[3], "No space targeted.")

	ui = callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[2], "my-organization")
	assert.NotContains(t, ui.Outputs[3], "my-space")
}

func TestTargetOrganizationWhenUserDoesNotHaveAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	configRepo.Login()

	orgs := []cf.Organization{}
	orgRepo.Organizations = orgs
	orgRepo.FindByNameErr = true

	ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[2], "No org targeted.")

	ui = callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")

	ui = callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[2], "No org targeted.")
}

func TestTargetOrganizationWhenOrgNotFound(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	configRepo.Delete()
	configRepo.Login()

	config, err := configRepo.Get()
	assert.NoError(t, err)
	org := cf.Organization{Guid: "previous-org-guid", Name: "previous-org"}
	config.Organization = org
	err = configRepo.Save()
	assert.NoError(t, err)

	orgRepo.DidNotFindOrganizationByName = true

	ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

	println(ui.DumpOutputs())
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "my-organization")
	assert.Contains(t, ui.Outputs[1], "not found")
}

// End test with organization option

// Start test with space option

func TestTargetSpaceWhenNoOrganizationIsSelected(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	configRepo.Login()

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Organization must be set before targeting space.")

	ui = callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[2], "No org targeted.")
}

func TestTargetSpaceWhenUserHasAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}

	spaces := []cf.Space{
		cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}
	spaceRepo.Spaces = spaces
	spaceRepo.SpaceByName = spaces[0]

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, spaceRepo.SpaceName, "my-space")
	assert.Contains(t, ui.Outputs[3], "space:")
	assert.Contains(t, ui.Outputs[3], "my-space")

	ui = callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[3], "my-space")
}

func TestTargetSpaceWhenUserDoesNotHaveAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}

	spaceRepo.SpaceByNameErr = true

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "You do not have access to that space.")

	ui = callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[3], "No space targeted.")
}

func TestTargetSpacehenSpaceNotFound(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.Organization = cf.Organization{Name: "my-org", Guid: "my-org-guid"}

	spaceRepo.DidNotFindSpaceByName = true

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "my-space")
	assert.Contains(t, ui.Outputs[1], "not found")
}

// End test with space option

func callTarget(args []string, reqFactory *testhelpers.FakeReqFactory,
	configRepo configuration.ConfigurationRepository, orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (ui *testhelpers.FakeUI) {

	ui = new(testhelpers.FakeUI)
	cmd := NewTarget(ui, configRepo, orgRepo, spaceRepo)
	ctxt := testhelpers.NewContext("target", args)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
