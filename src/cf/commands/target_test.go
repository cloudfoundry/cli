package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func getTargetDependencies() (orgRepo *testapi.FakeOrgRepository,
	spaceRepo *testapi.FakeSpaceRepository,
	configRepo *testconfig.FakeConfigRepository,
	reqFactory *testreq.FakeReqFactory) {

	orgRepo = &testapi.FakeOrgRepository{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	configRepo = &testconfig.FakeConfigRepository{}
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	return
}

func TestTargetFailsWithUsage(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callTarget([]string{"foo"}, reqFactory, configRepo, orgRepo, spaceRepo)
	assert.True(t, ui.FailedWithUsage)
}

func TestTargetRequirements(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	reqFactory.LoginSuccess = true

	callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestTargetOrganizationWhenUserHasAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Login()
	config, err := configRepo.Get()
	assert.NoError(t, err)

	config.SpaceFields = cf.SpaceFields{}
	config.SpaceFields.Name = "my-space"
	config.SpaceFields.Guid = "my-space-guid"

	org := cf.Organization{}
	org.Name = "my-organization"
	org.Guid = "my-organization-guid"

	orgRepo.Organizations = []cf.Organization{org}
	orgRepo.FindByNameOrganization = org

	ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, orgRepo.FindByNameName, "my-organization")
	assert.True(t, ui.ShowConfigurationCalled)

	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.OrganizationFields.Guid, "my-organization-guid")
}

func TestTargetOrganizationWhenUserDoesNotHaveAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	configRepo.Login()

	orgs := []cf.Organization{}
	orgRepo.Organizations = orgs
	orgRepo.FindByNameErr = true

	ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{{"FAILED"}})
}

func TestTargetOrganizationWhenOrgNotFound(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	configRepo.Delete()
	configRepo.Login()

	config, err := configRepo.Get()
	assert.NoError(t, err)

	config.OrganizationFields = cf.OrganizationFields{}
	config.OrganizationFields.Guid = "previous-org-guid"
	config.OrganizationFields.Name = "previous-org"

	err = configRepo.Save()
	assert.NoError(t, err)

	orgRepo.FindByNameNotFound = true

	ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"my-organization", "not found"},
	})
}

func TestTargetSpaceWhenNoOrganizationIsSelected(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	configRepo.Login()

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"An org must be targeted before targeting a space"},
	})
	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.OrganizationFields.Guid, "")
}

func TestTargetSpaceWhenUserHasAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.OrganizationFields = cf.OrganizationFields{}
	config.OrganizationFields.Name = "my-org"
	config.OrganizationFields.Guid = "my-org-guid"

	space := cf.Space{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"

	spaceRepo.Spaces = []cf.Space{space}
	spaceRepo.FindByNameSpace = space

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	assert.Equal(t, spaceRepo.FindByNameName, "my-space")
	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.SpaceFields.Guid, "my-space-guid")
	assert.True(t, ui.ShowConfigurationCalled)
}

func TestTargetSpaceWhenUserDoesNotHaveAccess(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.OrganizationFields = cf.OrganizationFields{}
	config.OrganizationFields.Name = "my-org"
	config.OrganizationFields.Guid = "my-org-guid"

	spaceRepo.FindByNameErr = true

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Unable to access space", "my-space"},
	})

	savedConfig := testconfig.SavedConfiguration
	assert.Equal(t, savedConfig.SpaceFields.Guid, "")
	assert.True(t, ui.ShowConfigurationCalled)
}

func TestTargetSpaceWhenSpaceNotFound(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

	configRepo.Delete()
	config := configRepo.Login()
	config.OrganizationFields = cf.OrganizationFields{}
	config.OrganizationFields.Name = "my-org"
	config.OrganizationFields.Guid = "my-org-guid"

	spaceRepo.FindByNameNotFound = true

	ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"my-space", "not found"},
	})
}

func TestTargetOrganizationAndSpace(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	configRepo.Delete()
	configRepo.Login()

	org := cf.Organization{}
	org.Name = "my-organization"
	org.Guid = "my-organization-guid"
	orgRepo.Organizations = []cf.Organization{org}

	space := cf.Space{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"
	spaceRepo.Spaces = []cf.Space{space}

	ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	savedConfig := testconfig.SavedConfiguration
	assert.True(t, ui.ShowConfigurationCalled)

	assert.Equal(t, orgRepo.FindByNameName, "my-organization")
	assert.Equal(t, savedConfig.OrganizationFields.Guid, "my-organization-guid")

	assert.Equal(t, spaceRepo.FindByNameName, "my-space")
	assert.Equal(t, savedConfig.SpaceFields.Guid, "my-space-guid")
}

func TestTargetOrganizationAndSpaceWhenSpaceFails(t *testing.T) {
	orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
	configRepo.Delete()
	configRepo.Login()

	org := cf.Organization{}
	org.Name = "my-organization"
	org.Guid = "my-organization-guid"
	orgRepo.Organizations = []cf.Organization{org}

	spaceRepo.FindByNameErr = true

	ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

	savedConfig := testconfig.SavedConfiguration
	assert.True(t, ui.ShowConfigurationCalled)

	assert.Equal(t, orgRepo.FindByNameName, "my-organization")
	assert.Equal(t, savedConfig.OrganizationFields.Guid, "my-organization-guid")
	assert.Equal(t, spaceRepo.FindByNameName, "my-space")
	assert.Equal(t, savedConfig.SpaceFields.Guid, "")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Unable to access space", "my-space"},
	})
}

func callTarget(args []string,
	reqFactory *testreq.FakeReqFactory,
	configRepo configuration.ConfigurationRepository,
	orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {

	ui = new(testterm.FakeUI)
	cmd := NewTarget(ui, configRepo, orgRepo, spaceRepo)
	ctxt := testcmd.NewContext("target", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
