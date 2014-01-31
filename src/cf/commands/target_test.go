package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestTargetFailsWithUsage", func() {
			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

			ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)
			assert.False(mr.T(), ui.FailedWithUsage)

			ui = callTarget([]string{"foo"}, reqFactory, configRepo, orgRepo, spaceRepo)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
		It("TestTargetRequirements", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
			reqFactory.LoginSuccess = true

			callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestTargetWithoutArgumentAndLoggedIn", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

			config := configRepo.Login()
			config.Target = "https://api.run.pivotal.io"

			ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

			assert.Equal(mr.T(), len(ui.Outputs), 2)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"No org targeted"},
				{"No space targeted"},
			})
		})
		It("TestTargetOrganizationWhenUserHasAccess", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

			configRepo.Login()
			config, err := configRepo.Get()
			assert.NoError(mr.T(), err)

			config.SpaceFields = cf.SpaceFields{}
			config.SpaceFields.Name = "my-space"
			config.SpaceFields.Guid = "my-space-guid"

			org := cf.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"

			orgRepo.Organizations = []cf.Organization{org}
			orgRepo.FindByNameOrganization = org

			ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

			assert.Equal(mr.T(), orgRepo.FindByNameName, "my-organization")
			assert.True(mr.T(), ui.ShowConfigurationCalled)

			savedConfig := testconfig.SavedConfiguration
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-organization-guid")
		})
		It("TestTargetOrganizationWhenUserDoesNotHaveAccess", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

			configRepo.Delete()
			configRepo.Login()

			orgs := []cf.Organization{}
			orgRepo.Organizations = orgs
			orgRepo.FindByNameErr = true

			ui := callTarget([]string{}, reqFactory, configRepo, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"No org targeted"},
			})

			ui = callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{{"FAILED"}})
		})
		It("TestTargetOrganizationWhenOrgNotFound", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
			configRepo.Delete()
			configRepo.Login()

			config, err := configRepo.Get()
			assert.NoError(mr.T(), err)

			config.OrganizationFields = cf.OrganizationFields{}
			config.OrganizationFields.Guid = "previous-org-guid"
			config.OrganizationFields.Name = "previous-org"

			err = configRepo.Save()
			assert.NoError(mr.T(), err)

			orgRepo.FindByNameNotFound = true

			ui := callTarget([]string{"-o", "my-organization"}, reqFactory, configRepo, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"my-organization", "not found"},
			})
		})
		It("TestTargetSpaceWhenNoOrganizationIsSelected", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

			configRepo.Delete()
			configRepo.Login()

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"An org must be targeted before targeting a space"},
			})
			savedConfig := testconfig.SavedConfiguration
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "")
		})
		It("TestTargetSpaceWhenUserHasAccess", func() {

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

			assert.Equal(mr.T(), spaceRepo.FindByNameName, "my-space")
			savedConfig := testconfig.SavedConfiguration
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.True(mr.T(), ui.ShowConfigurationCalled)
		})
		It("TestTargetSpaceWhenUserDoesNotHaveAccess", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

			configRepo.Delete()
			config := configRepo.Login()
			config.OrganizationFields = cf.OrganizationFields{}
			config.OrganizationFields.Name = "my-org"
			config.OrganizationFields.Guid = "my-org-guid"

			spaceRepo.FindByNameErr = true

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Unable to access space", "my-space"},
			})

			savedConfig := testconfig.SavedConfiguration
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "")
			assert.True(mr.T(), ui.ShowConfigurationCalled)
		})
		It("TestTargetSpaceWhenSpaceNotFound", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()

			configRepo.Delete()
			config := configRepo.Login()
			config.OrganizationFields = cf.OrganizationFields{}
			config.OrganizationFields.Name = "my-org"
			config.OrganizationFields.Guid = "my-org-guid"

			spaceRepo.FindByNameNotFound = true

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"my-space", "not found"},
			})
		})
		It("TestTargetOrganizationAndSpace", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
			configRepo.Delete()
			configRepo.Login()

			org := cf.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"
			orgRepo.FindByNameOrganization = org

			space := cf.Space{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"
			spaceRepo.FindByNameSpace = space

			ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

			savedConfig := testconfig.SavedConfiguration
			assert.True(mr.T(), ui.ShowConfigurationCalled)

			assert.Equal(mr.T(), orgRepo.FindByNameName, "my-organization")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-organization-guid")

			assert.Equal(mr.T(), spaceRepo.FindByNameName, "my-space")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
		})
		It("TestTargetOrganizationAndSpaceWhenSpaceFails", func() {

			orgRepo, spaceRepo, configRepo, reqFactory := getTargetDependencies()
			configRepo.Delete()
			configRepo.Login()

			org := cf.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"
			orgRepo.FindByNameOrganization = org

			spaceRepo.FindByNameErr = true

			ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, configRepo, orgRepo, spaceRepo)

			savedConfig := testconfig.SavedConfiguration
			assert.True(mr.T(), ui.ShowConfigurationCalled)

			assert.Equal(mr.T(), orgRepo.FindByNameName, "my-organization")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-organization-guid")
			assert.Equal(mr.T(), spaceRepo.FindByNameName, "my-space")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Unable to access space", "my-space"},
			})
		})
	})
}
