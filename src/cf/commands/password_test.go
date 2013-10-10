package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestPasswordRequiresValidAccessToken(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{ValidAccessTokenSuccess: false}
	configRepo := &testconfig.FakeConfigRepository{}
	callPassword([]string{}, reqFactory, &testapi.FakePasswordRepo{}, configRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{ValidAccessTokenSuccess: true}
	callPassword([]string{"", "", ""}, reqFactory, &testapi.FakePasswordRepo{}, configRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestPasswordCanBeChanged(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{ValidAccessTokenSuccess: true}
	pwdRepo := &testapi.FakePasswordRepo{Score: "meh"}
	configRepo := &testconfig.FakeConfigRepository{}
	ui := callPassword([]string{"old-password", "new-password", "new-password"}, reqFactory, pwdRepo, configRepo)

	assert.Contains(t, ui.PasswordPrompts[0], "Current Password")
	assert.Contains(t, ui.PasswordPrompts[1], "New Password")
	assert.Contains(t, ui.PasswordPrompts[2], "Verify Password")

	assert.Equal(t, pwdRepo.ScoredPassword, "new-password")
	assert.Contains(t, ui.Outputs[0], "Your password strength is: meh")

	assert.Contains(t, ui.Outputs[1], "Changing password...")
	assert.Equal(t, pwdRepo.UpdateNewPassword, "new-password")
	assert.Equal(t, pwdRepo.UpdateOldPassword, "old-password")
	assert.Contains(t, ui.Outputs[2], "OK")

	assert.Contains(t, ui.Outputs[3], "Please log in again")

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)
	assert.Empty(t, updatedConfig.AccessToken)
	assert.Equal(t, updatedConfig.Organization, cf.Organization{})
	assert.Equal(t, updatedConfig.Space, cf.Space{})
}

func TestPasswordVerification(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{ValidAccessTokenSuccess: true}
	pwdRepo := &testapi.FakePasswordRepo{Score: "meh"}
	configRepo := &testconfig.FakeConfigRepository{}
	ui := callPassword([]string{"old-password", "new-password", "new-password-with-error"}, reqFactory, pwdRepo, configRepo)

	assert.Contains(t, ui.PasswordPrompts[0], "Current Password")
	assert.Contains(t, ui.PasswordPrompts[1], "New Password")
	assert.Contains(t, ui.PasswordPrompts[2], "Verify Password")

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Password verification does not match")

	assert.Equal(t, pwdRepo.UpdateNewPassword, "")
}

func TestWhenCurrentPasswordDoesNotMatch(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{ValidAccessTokenSuccess: true}
	pwdRepo := &testapi.FakePasswordRepo{UpdateUnauthorized: true, Score: "meh"}
	configRepo := &testconfig.FakeConfigRepository{}
	ui := callPassword([]string{"old-password", "new-password", "new-password"}, reqFactory, pwdRepo, configRepo)

	assert.Contains(t, ui.PasswordPrompts[0], "Current Password")
	assert.Contains(t, ui.PasswordPrompts[1], "New Password")
	assert.Contains(t, ui.PasswordPrompts[2], "Verify Password")

	assert.Equal(t, pwdRepo.ScoredPassword, "new-password")
	assert.Contains(t, ui.Outputs[0], "Your password strength is: meh")

	assert.Contains(t, ui.Outputs[1], "Changing password...")
	assert.Equal(t, pwdRepo.UpdateNewPassword, "new-password")
	assert.Equal(t, pwdRepo.UpdateOldPassword, "old-password")
	assert.Contains(t, ui.Outputs[2], "FAILED")
	assert.Contains(t, ui.Outputs[3], "Current password did not match")
}

func callPassword(inputs []string, reqFactory *testreq.FakeReqFactory, pwdRepo *testapi.FakePasswordRepo, configRepo *testconfig.FakeConfigRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{Inputs: inputs}

	ctxt := testcmd.NewContext("passwd", []string{})
	cmd := NewPassword(ui, pwdRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
