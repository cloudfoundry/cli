package organization_test

import (
	. "cf/commands/organization"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callShowOrg(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("org", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        spaceFields,
		OrganizationFields: orgFields,
		AccessToken:        token,
	}

	cmd := NewShowOrg(ui, config)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestShowOrgRequirements", func() {
			args := []string{"my-org"}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			callShowOrg(mr.T(), args, reqFactory)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			callShowOrg(mr.T(), args, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestShowOrgFailsWithUsage", func() {

			org := models.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			reqFactory := &testreq.FakeReqFactory{Organization: org, LoginSuccess: true}

			args := []string{"my-org"}
			ui := callShowOrg(mr.T(), args, reqFactory)
			assert.False(mr.T(), ui.FailedWithUsage)

			args = []string{}
			ui = callShowOrg(mr.T(), args, reqFactory)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
		It("TestRunWhenOrganizationExists", func() {

			developmentSpaceFields := models.SpaceFields{}
			developmentSpaceFields.Name = "development"
			stagingSpaceFields := models.SpaceFields{}
			stagingSpaceFields.Name = "staging"
			domainFields := models.DomainFields{}
			domainFields.Name = "cfapps.io"
			cfAppDomainFields := models.DomainFields{}
			cfAppDomainFields.Name = "cf-app.com"
			org := models.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			org.QuotaDefinition = models.NewQuotaFields("cantina-quota", 512)
			org.Spaces = []models.SpaceFields{developmentSpaceFields, stagingSpaceFields}
			org.Domains = []models.DomainFields{domainFields, cfAppDomainFields}

			reqFactory := &testreq.FakeReqFactory{Organization: org, LoginSuccess: true}

			args := []string{"my-org"}
			ui := callShowOrg(mr.T(), args, reqFactory)

			assert.Equal(mr.T(), reqFactory.OrganizationName, "my-org")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting info for org", "my-org", "my-user"},
				{"OK"},
				{"my-org"},
				{"  domains:", "cfapps.io", "cf-app.com"},
				{"  quota: ", "cantina-quota", "512M"},
				{"  spaces:", "development", "staging"},
			})
		})
	})
}
