package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"strconv"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	"testhelpers/maker"
	testterm "testhelpers/terminal"
)

type LoginTestContext struct {
	Flags  []string
	Config *configuration.Configuration

	ui           *testterm.FakeUI
	authRepo     *testapi.FakeAuthenticationRepository
	endpointRepo *testapi.FakeEndpointRepo
	orgRepo      *testapi.FakeOrgRepository
	spaceRepo    *testapi.FakeSpaceRepository
}

func setUpLoginTestContext() (c *LoginTestContext) {
	c = new(LoginTestContext)
	c.Config = &configuration.Configuration{}

	c.ui = &testterm.FakeUI{}

	c.authRepo = &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		Config:       c.Config,
	}
	c.endpointRepo = &testapi.FakeEndpointRepo{Config: c.Config}

	org := models.Organization{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	c.orgRepo = &testapi.FakeOrgRepository{
		Organizations: []models.Organization{org},
	}

	space := models.Space{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"

	c.spaceRepo = &testapi.FakeSpaceRepository{
		Spaces: []models.Space{space},
	}

	return
}

func callLogin(c *LoginTestContext) {
	l := NewLogin(c.ui, c.Config, c.authRepo, c.endpointRepo, c.orgRepo, c.spaceRepo)
	testcmd.RunCommand(l, testcmd.NewContext("login", c.Flags), nil)
}

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSuccessfullyLoggingInWithNumericalPrompts", func() {
			c := setUpLoginTestContext()

			OUT_OF_RANGE_CHOICE := "3"

			c.ui.Inputs = []string{"api.example.com", "user@example.com", "password", OUT_OF_RANGE_CHOICE, "2", OUT_OF_RANGE_CHOICE, "1"}

			org1 := models.Organization{}
			org1.Guid = "some-org-guid"
			org1.Name = "some-org"

			org2 := models.Organization{}
			org2.Guid = "my-org-guid"
			org2.Name = "my-org"

			space1 := models.Space{}
			space1.Guid = "my-space-guid"
			space1.Name = "my-space"

			space2 := models.Space{}
			space2.Guid = "some-space-guid"
			space2.Name = "some-space"

			c.orgRepo.Organizations = []models.Organization{org1, org2}
			c.spaceRepo.Spaces = []models.Space{space1, space2}

			callLogin(c)

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Select an org"},
				{"1. some-org"},
				{"2. my-org"},
				{"Select a space"},
				{"1. my-space"},
				{"2. some-space"},
			})

			assert.Equal(mr.T(), c.Config.Target, "api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.Equal(mr.T(), c.orgRepo.FindByNameName, "my-org")
			assert.Equal(mr.T(), c.spaceRepo.FindByNameName, "my-space")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})

		It("TestSuccessfullyLoggingInWithStringPrompts", func() {
			c := setUpLoginTestContext()

			c.ui.Inputs = []string{"api.example.com", "user@example.com", "password", "my-org", "my-space"}

			org1 := models.Organization{}
			org1.Guid = "some-org-guid"
			org1.Name = "some-org"

			org2 := models.Organization{}
			org2.Guid = "my-org-guid"
			org2.Name = "my-org"

			space1 := models.Space{}
			space1.Guid = "my-space-guid"
			space1.Name = "my-space"

			space2 := models.Space{}
			space2.Guid = "some-space-guid"
			space2.Name = "some-space"

			c.orgRepo.Organizations = []models.Organization{org1, org2}
			c.spaceRepo.Spaces = []models.Space{space1, space2}

			callLogin(c)

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Select an org"},
				{"1. some-org"},
				{"2. my-org"},
				{"Select a space"},
				{"1. my-space"},
				{"2. some-space"},
			})

			assert.Equal(mr.T(), c.Config.Target, "api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.Equal(mr.T(), c.orgRepo.FindByNameName, "my-org")
			assert.Equal(mr.T(), c.spaceRepo.FindByNameName, "my-space")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})

		It("TestLoggingInWithTooManyOrgsDoesNotShowOrgList", func() {
			c := setUpLoginTestContext()

			c.ui.Inputs = []string{"api.example.com", "user@example.com", "password", "my-org-1", "my-space"}

			for i := 0; i < 60; i++ {
				id := strconv.Itoa(i)
				org := models.Organization{}
				org.Guid = "my-org-guid-" + id
				org.Name = "my-org-" + id
				c.orgRepo.Organizations = append(c.orgRepo.Organizations, org)
			}

			c.orgRepo.FindByNameOrganization = c.orgRepo.Organizations[1]

			space1 := models.Space{}
			space1.Guid = "my-space-guid"
			space1.Name = "my-space"

			space2 := models.Space{}
			space2.Guid = "some-space-guid"
			space2.Name = "some-space"

			c.spaceRepo.Spaces = []models.Space{space1, space2}

			callLogin(c)

			testassert.SliceDoesNotContain(mr.T(), c.ui.Outputs, testassert.Lines{
				{"my-org-2"},
			})
			assert.Equal(mr.T(), c.orgRepo.FindByNameName, "my-org-1")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid-1")
		})

		It("TestSuccessfullyLoggingInWithFlags", func() {
			c := setUpLoginTestContext()

			c.Flags = []string{"-a", "api.example.com", "-u", "user@example.com", "-p", "password", "-o", "my-org", "-s", "my-space"}

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})

		It("TestSuccessfullyLoggingInWithEndpointSetInConfig", func() {
			c := setUpLoginTestContext()

			c.Flags = []string{"-o", "my-org", "-s", "my-space"}
			c.ui.Inputs = []string{"user@example.com", "password"}
			c.Config.Target = "http://api.example.com"

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "http://api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})

		It("TestSuccessfullyLoggingInWithOrgSetInConfig", func() {
			c := setUpLoginTestContext()

			org := models.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			c.Config.OrganizationFields = org

			c.Flags = []string{"-s", "my-space"}
			c.ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}

			c.orgRepo.FindByNameOrganization = models.Organization{}

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "http://api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})

		It("TestSuccessfullyLoggingInWithOrgAndSpaceSetInConfig", func() {
			c := setUpLoginTestContext()

			org := models.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			c.Config.OrganizationFields = org

			space := models.SpaceFields{}
			space.Guid = "my-space-guid"
			space.Name = "my-space"
			c.Config.SpaceFields = space

			c.ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}

			c.orgRepo.FindByNameOrganization = models.Organization{}
			c.spaceRepo.FindByNameInOrgSpace = models.Space{}

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "http://api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})

		It("TestSuccessfullyLoggingInWithOnlyOneOrg", func() {
			c := setUpLoginTestContext()

			org := models.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			c.Flags = []string{"-s", "my-space"}
			c.ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}
			c.orgRepo.FindByNameOrganization = models.Organization{}
			c.orgRepo.Organizations = []models.Organization{org}

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "http://api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})

		It("TestSuccessfullyLoggingInWithOnlyOneSpace", func() {
			c := setUpLoginTestContext()

			space := models.Space{}
			space.Guid = "my-space-guid"
			space.Name = "my-space"

			c.Flags = []string{"-o", "my-org"}
			c.ui.Inputs = []string{"http://api.example.com", "user@example.com", "password"}
			c.spaceRepo.Spaces = []models.Space{space}

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "http://api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})

		It("TestUnsuccessfullyLoggingInWithAuthError", func() {
			c := setUpLoginTestContext()

			c.Flags = []string{"-u", "user@example.com"}
			c.ui.Inputs = []string{"api.example.com", "password", "password2", "password3"}
			c.authRepo.AuthError = true

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "api.example.com")
			assert.Empty(mr.T(), c.Config.OrganizationFields.Guid)
			assert.Empty(mr.T(), c.Config.SpaceFields.Guid)
			assert.Empty(mr.T(), c.Config.AccessToken)
			assert.Empty(mr.T(), c.Config.RefreshToken)

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})

		It("TestUnsuccessfullyLoggingInWithUpdateEndpointError", func() {
			c := setUpLoginTestContext()

			c.ui.Inputs = []string{"api.example.com"}
			c.endpointRepo.UpdateEndpointError = net.NewApiResponseWithMessage("Server error")

			callLogin(c)

			assert.Empty(mr.T(), c.Config.Target)
			assert.Empty(mr.T(), c.Config.OrganizationFields.Guid)
			assert.Empty(mr.T(), c.Config.SpaceFields.Guid)
			assert.Empty(mr.T(), c.Config.AccessToken)
			assert.Empty(mr.T(), c.Config.RefreshToken)

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})

		It("TestUnsuccessfullyLoggingInWithOrgFindByNameErr", func() {
			c := setUpLoginTestContext()

			c.Flags = []string{"-u", "user@example.com", "-o", "my-org", "-s", "my-space"}
			c.ui.Inputs = []string{"api.example.com", "user@example.com", "password"}
			c.orgRepo.FindByNameErr = true

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "api.example.com")
			assert.Empty(mr.T(), c.Config.OrganizationFields.Guid)
			assert.Empty(mr.T(), c.Config.SpaceFields.Guid)
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})

		It("TestUnsuccessfullyLoggingInWithSpaceFindByNameErr", func() {
			c := setUpLoginTestContext()

			c.Flags = []string{"-u", "user@example.com", "-o", "my-org", "-s", "my-space"}
			c.ui.Inputs = []string{"api.example.com", "user@example.com", "password"}
			c.spaceRepo.FindByNameErr = true

			callLogin(c)

			assert.Equal(mr.T(), c.Config.Target, "api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "my-org-guid")
			assert.Empty(mr.T(), c.Config.SpaceFields.Guid)
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})

		It("TestSuccessfullyLoggingInWithoutTargetOrg", func() {
			c := setUpLoginTestContext()

			c.ui.Inputs = []string{"api.example.com", "user@example.com", "password", ""}

			org1 := maker.NewOrg(maker.Overrides{"name": "org1"})
			org2 := maker.NewOrg(maker.Overrides{"name": "org2"})
			c.orgRepo.Organizations = []models.Organization{org1, org2}

			callLogin(c)

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Select an org (or press enter to skip):"},
			})
			testassert.SliceDoesNotContain(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Select a space", "or press enter to skip"},
			})
			assert.Equal(mr.T(), c.Config.Target, "api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")
		})

		It("TestSuccessfullyLoggingInWithoutTargetSpace", func() {
			c := setUpLoginTestContext()

			c.ui.Inputs = []string{"api.example.com", "user@example.com", "password", ""}

			org := models.Organization{}
			org.Guid = "some-org-guid"
			org.Name = "some-org"

			space1 := maker.NewSpace(maker.Overrides{"name": "some-space", "guid": "some-space-guid"})
			space2 := maker.NewSpace(maker.Overrides{"name": "other-space", "guid": "other-space-guid"})

			c.orgRepo.Organizations = []models.Organization{org}
			c.spaceRepo.Spaces = []models.Space{space1, space2}

			callLogin(c)

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Select a space (or press enter to skip):"},
			})
			testassert.SliceDoesNotContain(mr.T(), c.ui.Outputs, testassert.Lines{
				{"FAILED"},
			})

			assert.Equal(mr.T(), c.Config.Target, "api.example.com")
			assert.Equal(mr.T(), c.Config.OrganizationFields.Guid, "some-org-guid")
			assert.Equal(mr.T(), c.Config.SpaceFields.Guid, "")
			assert.Equal(mr.T(), c.Config.AccessToken, "my_access_token")
			assert.Equal(mr.T(), c.Config.RefreshToken, "my_refresh_token")
		})
	})
}
