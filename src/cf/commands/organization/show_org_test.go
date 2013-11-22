package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
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
	org.Spaces = []cf.SpaceFields{developmentSpaceFields, stagingSpaceFields}
	org.Domains = []cf.DomainFields{domainFields, cfAppDomainFields}

	reqFactory := &testreq.FakeReqFactory{Organization: org, LoginSuccess: true}

	args := []string{"my-org"}
	ui := callShowOrg(t, args, reqFactory)

	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	assert.Equal(t, len(ui.Outputs), 5)
	assert.Contains(t, ui.Outputs[0], "Getting info for org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "my-org")
	assert.Contains(t, ui.Outputs[3], "  domains:")
	assert.Contains(t, ui.Outputs[3], "cfapps.io, cf-app.com")
	assert.Contains(t, ui.Outputs[4], "  spaces:")
	assert.Contains(t, ui.Outputs[4], "development, staging")
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
