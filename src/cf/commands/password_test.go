package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestPasswordRequiresValidAccessToken(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{ValidAccessTokenSuccess: false}
	configRepo := &testhelpers.FakeConfigRepository{}
	callPassword([]string{}, reqFactory, &testhelpers.FakePasswordRepo{}, configRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{ValidAccessTokenSuccess: true}
	callPassword([]string{"", "", ""}, reqFactory, &testhelpers.FakePasswordRepo{}, configRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
}

func TestPasswordCanBeChanged(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{ValidAccessTokenSuccess: true}
	pwdRepo := &testhelpers.FakePasswordRepo{Score: "meh"}
	configRepo := &testhelpers.FakeConfigRepository{}
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

	assert.Contains(t, ui.Outputs[3], "Please log back in.")

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)
	assert.Empty(t, updatedConfig.AccessToken)
	assert.Equal(t, updatedConfig.Organization, cf.Organization{})
	assert.Equal(t, updatedConfig.Space, cf.Space{})
}

func TestPasswordVerification(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{ValidAccessTokenSuccess: true}
	pwdRepo := &testhelpers.FakePasswordRepo{Score: "meh"}
	configRepo := &testhelpers.FakeConfigRepository{}
	ui := callPassword([]string{"old-password", "new-password", "new-password-with-error"}, reqFactory, pwdRepo, configRepo)

	assert.Contains(t, ui.PasswordPrompts[0], "Current Password")
	assert.Contains(t, ui.PasswordPrompts[1], "New Password")
	assert.Contains(t, ui.PasswordPrompts[2], "Verify Password")

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Password verification does not match")

	assert.Equal(t, pwdRepo.UpdateNewPassword, "")
}

func TestWhenCurrentPasswordDoesNotMatch(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{ValidAccessTokenSuccess: true}
	pwdRepo := &testhelpers.FakePasswordRepo{UpdateUnauthorized: true, Score: "meh"}
	configRepo := &testhelpers.FakeConfigRepository{}
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

func callPassword(inputs []string, reqFactory *testhelpers.FakeReqFactory, pwdRepo *testhelpers.FakePasswordRepo, configRepo *testhelpers.FakeConfigRepository) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{Inputs: inputs}

	ctxt := testhelpers.NewContext("passwd", []string{})
	cmd := NewPassword(ui, pwdRepo, configRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
