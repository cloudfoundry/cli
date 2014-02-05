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
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func getTargetDependencies() (
	orgRepo *testapi.FakeOrgRepository,
	spaceRepo *testapi.FakeSpaceRepository,
	config *configuration.Configuration,
	reqFactory *testreq.FakeReqFactory) {

	orgRepo = &testapi.FakeOrgRepository{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	config = &configuration.Configuration{}

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	return
}

func callTarget(args []string,
	reqFactory *testreq.FakeReqFactory,
	config *configuration.Configuration,
	orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {

	ui = new(testterm.FakeUI)
	cmd := NewTarget(ui, config, orgRepo, spaceRepo)
	ctxt := testcmd.NewContext("target", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func simulateLogin(c *configuration.Configuration) {
	c.AccessToken = `BEARER eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E`
}

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestTargetFailsWithUsage", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			ui := callTarget([]string{}, reqFactory, config, orgRepo, spaceRepo)
			assert.False(mr.T(), ui.FailedWithUsage)

			ui = callTarget([]string{"foo"}, reqFactory, config, orgRepo, spaceRepo)
			assert.True(mr.T(), ui.FailedWithUsage)
		})

		It("TestTargetRequirements", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()
			reqFactory.LoginSuccess = true

			callTarget([]string{}, reqFactory, config, orgRepo, spaceRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})

		It("TestTargetOrganizationWhenUserHasAccess", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)

			config.SpaceFields = cf.SpaceFields{}
			config.SpaceFields.Name = "my-space"
			config.SpaceFields.Guid = "my-space-guid"

			org := cf.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"

			orgRepo.Organizations = []cf.Organization{org}
			orgRepo.FindByNameOrganization = org

			ui := callTarget([]string{"-o", "my-organization"}, reqFactory, config, orgRepo, spaceRepo)

			assert.Equal(mr.T(), orgRepo.FindByNameName, "my-organization")
			assert.True(mr.T(), ui.ShowConfigurationCalled)

			assert.Equal(mr.T(), config.OrganizationFields.Guid, "my-organization-guid")
		})

		It("TestTargetOrganizationWhenUserDoesNotHaveAccess", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)

			orgs := []cf.Organization{}
			orgRepo.Organizations = orgs
			orgRepo.FindByNameErr = true

			ui := callTarget([]string{"-o", "my-organization"}, reqFactory, config, orgRepo, spaceRepo)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{{"FAILED"}})
		})

		It("TestTargetOrganizationWhenOrgNotFound", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)
			config.OrganizationFields = cf.OrganizationFields{}
			config.OrganizationFields.Guid = "previous-org-guid"
			config.OrganizationFields.Name = "previous-org"

			orgRepo.FindByNameNotFound = true

			ui := callTarget([]string{"-o", "my-organization"}, reqFactory, config, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"my-organization", "not found"},
			})
		})

		It("TestTargetSpaceWhenNoOrganizationIsSelected", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"An org must be targeted before targeting a space"},
			})
			assert.Equal(mr.T(), config.OrganizationFields.Guid, "")
		})

		It("TestTargetSpaceWhenUserHasAccess", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)
			config.OrganizationFields = cf.OrganizationFields{}
			config.OrganizationFields.Name = "my-org"
			config.OrganizationFields.Guid = "my-org-guid"

			space := cf.Space{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"

			spaceRepo.Spaces = []cf.Space{space}
			spaceRepo.FindByNameSpace = space

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			assert.Equal(mr.T(), spaceRepo.FindByNameName, "my-space")
			assert.Equal(mr.T(), config.SpaceFields.Guid, "my-space-guid")
			assert.True(mr.T(), ui.ShowConfigurationCalled)
		})

		It("TestTargetSpaceWhenUserDoesNotHaveAccess", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)
			config.OrganizationFields = cf.OrganizationFields{}
			config.OrganizationFields.Name = "my-org"
			config.OrganizationFields.Guid = "my-org-guid"

			spaceRepo.FindByNameErr = true

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Unable to access space", "my-space"},
			})

			assert.Equal(mr.T(), config.SpaceFields.Guid, "")
			assert.False(mr.T(), ui.ShowConfigurationCalled)
		})

		It("TestTargetSpaceWhenSpaceNotFound", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)
			config.OrganizationFields = cf.OrganizationFields{}
			config.OrganizationFields.Name = "my-org"
			config.OrganizationFields.Guid = "my-org-guid"

			spaceRepo.FindByNameNotFound = true

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"my-space", "not found"},
			})
		})

		It("TestTargetOrganizationAndSpace", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)

			org := cf.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"
			orgRepo.Organizations = []cf.Organization{org}

			space := cf.Space{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"
			spaceRepo.Spaces = []cf.Space{space}

			ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			assert.True(mr.T(), ui.ShowConfigurationCalled)
			assert.Equal(mr.T(), orgRepo.FindByNameName, "my-organization")
			assert.Equal(mr.T(), config.OrganizationFields.Guid, "my-organization-guid")
			assert.Equal(mr.T(), spaceRepo.FindByNameName, "my-space")
			assert.Equal(mr.T(), config.SpaceFields.Guid, "my-space-guid")
		})

		It("TestTargetOrganizationAndSpaceWhenSpaceFails", func() {
			orgRepo, spaceRepo, config, reqFactory := getTargetDependencies()

			simulateLogin(config)

			org := cf.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"
			orgRepo.Organizations = []cf.Organization{org}

			spaceRepo.FindByNameErr = true

			ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			assert.False(mr.T(), ui.ShowConfigurationCalled)
			assert.Equal(mr.T(), orgRepo.FindByNameName, "my-organization")
			assert.Equal(mr.T(), config.OrganizationFields.Guid, "my-organization-guid")
			assert.Equal(mr.T(), spaceRepo.FindByNameName, "my-space")
			assert.Equal(mr.T(), config.SpaceFields.Guid, "")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Unable to access space", "my-space"},
			})
		})
	})
}
