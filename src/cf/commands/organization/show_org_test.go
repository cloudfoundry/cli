package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestShowOrgRequirements(t *testing.T) {
	args := []string{"my-org"}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callShowOrg(t, args, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callShowOrg(t, args, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestShowOrgFailsWithUsage(t *testing.T) {
	org := cf.Organization{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"
	reqFactory := &testreq.FakeReqFactory{Organization: org, LoginSuccess: true}

	args := []string{"my-org"}
	ui := callShowOrg(t, args, reqFactory)
	assert.False(t, ui.FailedWithUsage)

	args = []string{}
	ui = callShowOrg(t, args, reqFactory)
	assert.True(t, ui.FailedWithUsage)
}

func TestRunWhenOrganizationExists(t *testing.T) {
	developmentSpaceFields := cf.SpaceFields{}
	developmentSpaceFields.Name = "development"
	stagingSpaceFields := cf.SpaceFields{}
	stagingSpaceFields.Name = "staging"
	domainFields := cf.DomainFields{}
	domainFields.Name = "cfapps.io"
	cfAppDomainFields := cf.DomainFields{}
	cfAppDomainFields.Name = "cf-app.com"
	org := cf.Organization{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"
	org.MemoryLimit = "256M"
	org.Spaces = []cf.SpaceFields{developmentSpaceFields, stagingSpaceFields}
	org.Domains = []cf.DomainFields{domainFields, cfAppDomainFields}

	reqFactory := &testreq.FakeReqFactory{Organization: org, LoginSuccess: true}

	args := []string{"my-org"}
	ui := callShowOrg(t, args, reqFactory)

	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting info for org", "my-org", "my-user"},
		{"OK"},
		{"my-org"},
		{"  domains:", "cfapps.io", "cf-app.com"},
		{"  quota: ", "256M"},
		{"  spaces:", "development", "staging"},
	})
}

func callShowOrg(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("org", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	spaceFields := cf.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := cf.OrganizationFields{}
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
