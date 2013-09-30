package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestShowOrgRequirements(t *testing.T) {
	args := []string{"my-org"}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callShowOrg(args, reqFactory)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}
	callShowOrg(args, reqFactory)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestShowOrgFailsWithUsage(t *testing.T) {
	org := cf.Organization{Name: "my-org", Guid: "my-org-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Organization: org, LoginSuccess: true}

	args := []string{"my-org"}
	ui := callShowOrg(args, reqFactory)
	assert.False(t, ui.FailedWithUsage)

	args = []string{}
	ui = callShowOrg(args, reqFactory)
	assert.True(t, ui.FailedWithUsage)
}

func TestRunWhenOrganizationExists(t *testing.T) {
	development := cf.Space{Name: "development"}
	staging := cf.Space{Name: "staging"}
	domain := cf.Domain{Name: "cfapps.io"}
	cfAppDomain := cf.Domain{Name: "cf-app.com"}
	org := cf.Organization{
		Name:    "my-org",
		Guid:    "my-org-guid",
		Spaces:  []cf.Space{development, staging},
		Domains: []cf.Domain{domain, cfAppDomain},
	}

	reqFactory := &testhelpers.FakeReqFactory{Organization: org, LoginSuccess: true}

	args := []string{"my-org"}
	ui := callShowOrg(args, reqFactory)

	assert.Equal(t, reqFactory.OrganizationName, "my-org")

	assert.Equal(t, len(ui.Outputs), 5)
	assert.Contains(t, ui.Outputs[0], "Getting info")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "my-org")
	assert.Contains(t, ui.Outputs[3], "  domains:")
	assert.Contains(t, ui.Outputs[3], "cfapps.io, cf-app.com")
	assert.Contains(t, ui.Outputs[4], "  spaces:")
	assert.Contains(t, ui.Outputs[4], "development, staging")
}

func callShowOrg(args []string, reqFactory *testhelpers.FakeReqFactory) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("org", args)

	cmd := NewShowOrg(ui)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
