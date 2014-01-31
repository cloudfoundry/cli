package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"strconv"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

type LoginTestContext struct {
	Flags  []string
	Inputs []string
	Config configuration.Configuration

	configRepo   testconfig.FakeConfigRepository
	ui           *testterm.FakeUI
	authRepo     *testapi.FakeAuthenticationRepository
	endpointRepo *testapi.FakeEndpointRepo
	orgRepo      *testapi.FakeOrgRepository
	spaceRepo    *testapi.FakeSpaceRepository
}

func defaultBeforeBlock(*LoginTestContext) {}

func callLogin(t mr.TestingT, c *LoginTestContext, beforeBlock func(*LoginTestContext)) {

	c.configRepo = testconfig.FakeConfigRepository{}
	c.ui = &testterm.FakeUI{
		Inputs: c.Inputs,
	}
	c.authRepo = &testapi.FakeAuthenticationRepository{
		AccessToken:  "my_access_token",
		RefreshToken: "my_refresh_token",
		ConfigRepo:   c.configRepo,
	}
	c.endpointRepo = &testapi.FakeEndpointRepo{}

	org := cf.Organization{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"

	c.orgRepo = &testapi.FakeOrgRepository{
		FindByNameOrganization: org,
	}

	space := cf.Space{}
	space.Name = "my-space"
	space.Guid = "my-space-guid"

	c.spaceRepo = &testapi.FakeSpaceRepository{
		FindByNameSpace: space,
	}

	c.configRepo.Delete()
	config, _ := c.configRepo.Get()
	config.Target = c.Config.Target
	config.OrganizationFields = c.Config.OrganizationFields
	config.SpaceFields = c.Config.SpaceFields

	beforeBlock(c)

	l := NewLogin(c.ui, c.configRepo, c.authRepo, c.endpointRepo, c.orgRepo, c.spaceRepo)
	l.Run(testcmd.NewContext("login", c.Flags))
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSuccessfullyLoggingInWithNumericalPrompts", func() {

			OUT_OF_RANGE_CHOICE := "3"
			c := LoginTestContext{
				Inputs: []string{"api.example.com", "user@example.com", "password", OUT_OF_RANGE_CHOICE, "2", OUT_OF_RANGE_CHOICE, "1"},
			}

			org1 := cf.Organization{}
			org1.Guid = "some-org-guid"
			org1.Name = "some-org"

			org2 := cf.Organization{}
			org2.Guid = "my-org-guid"
			org2.Name = "my-org"

			space1 := cf.Space{}
			space1.Guid = "my-space-guid"
			space1.Name = "my-space"

			space2 := cf.Space{}
			space2.Guid = "some-space-guid"
			space2.Name = "some-space"

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.orgRepo.Organizations = []cf.Organization{org1, org2}
				c.spaceRepo.Spaces = []cf.Space{space1, space2}
			})

			savedConfig := testconfig.SavedConfiguration

			expectedOutputs := testassert.Lines{
				testassert.Line{"Select an org:"},
				testassert.Line{"1. some-org"},
				testassert.Line{"2. my-org"},
				testassert.Line{"Select a space:"},
				testassert.Line{"1. my-space"},
				testassert.Line{"2. some-space"},
			}

			testassert.SliceContains(mr.T(), c.ui.Outputs, expectedOutputs)

			assert.Equal(mr.T(), savedConfig.Target, "api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.Equal(mr.T(), c.orgRepo.FindByNameName, "my-org")
			assert.Equal(mr.T(), c.spaceRepo.FindByNameName, "my-space")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})
		It("TestSuccessfullyLoggingInWithStringPrompts", func() {

			c := LoginTestContext{
				Inputs: []string{"api.example.com", "user@example.com", "password", "my-org", "my-space"},
			}

			org1 := cf.Organization{}
			org1.Guid = "some-org-guid"
			org1.Name = "some-org"

			org2 := cf.Organization{}
			org2.Guid = "my-org-guid"
			org2.Name = "my-org"

			space1 := cf.Space{}
			space1.Guid = "my-space-guid"
			space1.Name = "my-space"

			space2 := cf.Space{}
			space2.Guid = "some-space-guid"
			space2.Name = "some-space"

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.orgRepo.Organizations = []cf.Organization{org1, org2}
				c.spaceRepo.Spaces = []cf.Space{space1, space2}
			})

			savedConfig := testconfig.SavedConfiguration

			expectedOutputs := testassert.Lines{
				testassert.Line{"Select an org:"},
				testassert.Line{"1. some-org"},
				testassert.Line{"2. my-org"},
				testassert.Line{"Select a space:"},
				testassert.Line{"1. my-space"},
				testassert.Line{"2. some-space"},
			}

			testassert.SliceContains(mr.T(), c.ui.Outputs, expectedOutputs)

			assert.Equal(mr.T(), savedConfig.Target, "api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.Equal(mr.T(), c.orgRepo.FindByNameName, "my-org")
			assert.Equal(mr.T(), c.spaceRepo.FindByNameName, "my-space")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})
		It("TestLoggingInWithTooManyOrgsDoesNotShowOrgList", func() {

			c := LoginTestContext{
				Inputs: []string{"api.example.com", "user@example.com", "password", "my-org-1", "my-space"},
			}

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				for i := 0; i < 60; i++ {
					id := strconv.Itoa(i)
					org := cf.Organization{}
					org.Guid = "my-org-guid-" + id
					org.Name = "my-org-" + id
					c.orgRepo.Organizations = append(c.orgRepo.Organizations, org)
				}

				c.orgRepo.FindByNameOrganization = c.orgRepo.Organizations[1]

				space1 := cf.Space{}
				space1.Guid = "my-space-guid"
				space1.Name = "my-space"

				space2 := cf.Space{}
				space2.Guid = "some-space-guid"
				space2.Name = "some-space"

				c.spaceRepo.Spaces = []cf.Space{space1, space2}
			})

			savedConfig := testconfig.SavedConfiguration

			testassert.SliceDoesNotContain(mr.T(), c.ui.Outputs, testassert.Lines{
				{"my-org-2"},
			})
			assert.Equal(mr.T(), c.orgRepo.FindByNameName, "my-org-1")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid-1")
		})
		It("TestSuccessfullyLoggingInWithFlags", func() {

			c := LoginTestContext{
				Flags: []string{"-a", "api.example.com", "-u", "user@example.com", "-p", "password", "-o", "my-org", "-s", "my-space"},
			}

			callLogin(mr.T(), &c, defaultBeforeBlock)

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})
		It("TestSuccessfullyLoggingInWithEndpointSetInConfig", func() {

			existingConfig := configuration.Configuration{
				Target: "http://api.example.com",
			}

			c := LoginTestContext{
				Flags:  []string{"-o", "my-org", "-s", "my-space"},
				Inputs: []string{"user@example.com", "password"},
				Config: existingConfig,
			}

			callLogin(mr.T(), &c, defaultBeforeBlock)

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "http://api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})
		It("TestSuccessfullyLoggingInWithOrgSetInConfig", func() {

			org := cf.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			existingConfig := configuration.Configuration{OrganizationFields: org}

			c := LoginTestContext{
				Flags:  []string{"-s", "my-space"},
				Inputs: []string{"http://api.example.com", "user@example.com", "password"},
				Config: existingConfig,
			}

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.orgRepo.FindByNameOrganization = cf.Organization{}
			})

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "http://api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})
		It("TestSuccessfullyLoggingInWithOrgAndSpaceSetInConfig", func() {

			org := cf.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			space := cf.SpaceFields{}
			space.Guid = "my-space-guid"
			space.Name = "my-space"

			existingConfig := configuration.Configuration{
				OrganizationFields: org,
				SpaceFields:        space,
			}

			c := LoginTestContext{
				Inputs: []string{"http://api.example.com", "user@example.com", "password"},
				Config: existingConfig,
			}

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.orgRepo.FindByNameOrganization = cf.Organization{}
				c.spaceRepo.FindByNameInOrgSpace = cf.Space{}
			})

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "http://api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})
		It("TestSuccessfullyLoggingInWithOnlyOneOrg", func() {

			org := cf.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			c := LoginTestContext{
				Flags:  []string{"-s", "my-space"},
				Inputs: []string{"http://api.example.com", "user@example.com", "password"},
			}

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.orgRepo.FindByNameOrganization = cf.Organization{}
				c.orgRepo.Organizations = []cf.Organization{org}
			})

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "http://api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})
		It("TestSuccessfullyLoggingInWithOnlyOneSpace", func() {

			space := cf.Space{}
			space.Guid = "my-space-guid"
			space.Name = "my-space"

			c := LoginTestContext{
				Flags:  []string{"-o", "my-org"},
				Inputs: []string{"http://api.example.com", "user@example.com", "password"},
			}

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.spaceRepo.Spaces = []cf.Space{space}
			})

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "http://api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Equal(mr.T(), savedConfig.SpaceFields.Guid, "my-space-guid")
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			assert.Equal(mr.T(), c.endpointRepo.UpdateEndpointReceived, "http://api.example.com")
			assert.Equal(mr.T(), c.authRepo.Email, "user@example.com")
			assert.Equal(mr.T(), c.authRepo.Password, "password")

			assert.True(mr.T(), c.ui.ShowConfigurationCalled)
		})
		It("TestUnsuccessfullyLoggingInWithAuthError", func() {

			c := LoginTestContext{
				Flags:  []string{"-u", "user@example.com"},
				Inputs: []string{"api.example.com", "password", "password2", "password3"},
			}

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.authRepo.AuthError = true
			})

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "api.example.com")
			assert.Empty(mr.T(), savedConfig.OrganizationFields.Guid)
			assert.Empty(mr.T(), savedConfig.SpaceFields.Guid)
			assert.Empty(mr.T(), savedConfig.AccessToken)
			assert.Empty(mr.T(), savedConfig.RefreshToken)

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})
		It("TestUnsuccessfullyLoggingInWithUpdateEndpointError", func() {

			c := LoginTestContext{
				Inputs: []string{"api.example.com"},
			}
			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.endpointRepo.UpdateEndpointError = net.NewApiResponseWithMessage("Server error")
			})

			savedConfig := testconfig.SavedConfiguration

			assert.Empty(mr.T(), savedConfig.Target)
			assert.Empty(mr.T(), savedConfig.OrganizationFields.Guid)
			assert.Empty(mr.T(), savedConfig.SpaceFields.Guid)
			assert.Empty(mr.T(), savedConfig.AccessToken)
			assert.Empty(mr.T(), savedConfig.RefreshToken)

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})
		It("TestUnsuccessfullyLoggingInWithOrgFindByNameErr", func() {

			c := LoginTestContext{
				Flags:  []string{"-u", "user@example.com", "-o", "my-org", "-s", "my-space"},
				Inputs: []string{"api.example.com", "user@example.com", "password"},
			}

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.orgRepo.FindByNameErr = true
			})

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "api.example.com")
			assert.Empty(mr.T(), savedConfig.OrganizationFields.Guid)
			assert.Empty(mr.T(), savedConfig.SpaceFields.Guid)
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})
		It("TestUnsuccessfullyLoggingInWithSpaceFindByNameErr", func() {

			c := LoginTestContext{
				Flags:  []string{"-u", "user@example.com", "-o", "my-org", "-s", "my-space"},
				Inputs: []string{"api.example.com", "user@example.com", "password"},
			}

			callLogin(mr.T(), &c, func(c *LoginTestContext) {
				c.spaceRepo.FindByNameErr = true
			})

			savedConfig := testconfig.SavedConfiguration

			assert.Equal(mr.T(), savedConfig.Target, "api.example.com")
			assert.Equal(mr.T(), savedConfig.OrganizationFields.Guid, "my-org-guid")
			assert.Empty(mr.T(), savedConfig.SpaceFields.Guid)
			assert.Equal(mr.T(), savedConfig.AccessToken, "my_access_token")
			assert.Equal(mr.T(), savedConfig.RefreshToken, "my_refresh_token")

			testassert.SliceContains(mr.T(), c.ui.Outputs, testassert.Lines{
				{"Failed"},
			})
		})
	})
}
