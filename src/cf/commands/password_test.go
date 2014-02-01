package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestPasswordRequiresValidAccessToken(t *testing.T) {
	deps := getPasswordDeps()
	deps.ReqFactory.ValidAccessTokenSuccess = false
	callPassword([]string{}, deps)
	assert.False(t, testcmd.CommandDidPassRequirements)

	deps.ReqFactory.ValidAccessTokenSuccess = true
	deps.PwdRepo.UpdateUnauthorized = false

	callPassword([]string{"", "", ""}, deps)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestPasswordCanBeChanged(t *testing.T) {
	deps := getPasswordDeps()
	deps.ReqFactory.ValidAccessTokenSuccess = true
	deps.PwdRepo.UpdateUnauthorized = false
	ui := callPassword([]string{"old-password", "new-password", "new-password"}, deps)

	testassert.SliceContains(t, ui.PasswordPrompts, testassert.Lines{
		{"Current Password"},
		{"New Password"},
		{"Verify Password"},
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Changing password..."},
		{"OK"},
		{"Please log in again"},
	})

	assert.Equal(t, deps.PwdRepo.UpdateNewPassword, "new-password")
	assert.Equal(t, deps.PwdRepo.UpdateOldPassword, "old-password")

	updatedConfig, err := deps.ConfigRepo.Get()
	assert.NoError(t, err)
	assert.Empty(t, updatedConfig.AccessToken)
	assert.Equal(t, updatedConfig.OrganizationFields, cf.OrganizationFields{})
	assert.Equal(t, updatedConfig.SpaceFields, cf.SpaceFields{})
}

func TestPasswordVerification(t *testing.T) {
	deps := getPasswordDeps()
	deps.ReqFactory.ValidAccessTokenSuccess = true
	ui := callPassword([]string{"old-password", "new-password", "new-password-with-error"}, deps)

	testassert.SliceContains(t, ui.PasswordPrompts, testassert.Lines{
		{"Current Password"},
		{"New Password"},
		{"Verify Password"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Password verification does not match"},
	})

	assert.Equal(t, deps.PwdRepo.UpdateNewPassword, "")
}

func TestWhenCurrentPasswordDoesNotMatch(t *testing.T) {
	deps := getPasswordDeps()
	deps.ReqFactory.ValidAccessTokenSuccess = true
	deps.PwdRepo.UpdateUnauthorized = true
	ui := callPassword([]string{"old-password", "new-password", "new-password"}, deps)

	testassert.SliceContains(t, ui.PasswordPrompts, testassert.Lines{
		{"Current Password"},
		{"New Password"},
		{"Verify Password"},
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Changing password..."},
		{"FAILED"},
		{"Current password did not match"},
	})

	assert.Equal(t, deps.PwdRepo.UpdateNewPassword, "new-password")
	assert.Equal(t, deps.PwdRepo.UpdateOldPassword, "old-password")
}

type passwordDeps struct {
	ReqFactory *testreq.FakeReqFactory
	PwdRepo    *testapi.FakePasswordRepo
	ConfigRepo *testconfig.FakeConfigRepository
}

func getPasswordDeps() passwordDeps {
	return passwordDeps{
		ReqFactory: &testreq.FakeReqFactory{ValidAccessTokenSuccess: true},
		PwdRepo:    &testapi.FakePasswordRepo{UpdateUnauthorized: true},
		ConfigRepo: &testconfig.FakeConfigRepository{},
	}
}

func callPassword(inputs []string, deps passwordDeps) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{Inputs: inputs}

	ctxt := testcmd.NewContext("passwd", []string{})
	cmd := NewPassword(ui, deps.PwdRepo, deps.ConfigRepo)
	testcmd.RunCommand(cmd, ctxt, deps.ReqFactory)

	return
}
