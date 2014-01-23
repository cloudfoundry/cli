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
	pwdRepo := &testapi.FakePasswordRepo{}
	configRepo := &testconfig.FakeConfigRepository{}
	ui := callPassword([]string{"old-password", "new-password", "new-password"}, reqFactory, pwdRepo, configRepo)

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

	assert.Equal(t, pwdRepo.UpdateNewPassword, "new-password")
	assert.Equal(t, pwdRepo.UpdateOldPassword, "old-password")

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)
	assert.Empty(t, updatedConfig.AccessToken)
	assert.Equal(t, updatedConfig.OrganizationFields, cf.OrganizationFields{})
	assert.Equal(t, updatedConfig.SpaceFields, cf.SpaceFields{})
}

func TestPasswordVerification(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{ValidAccessTokenSuccess: true}
	pwdRepo := &testapi.FakePasswordRepo{}
	configRepo := &testconfig.FakeConfigRepository{}
	ui := callPassword([]string{"old-password", "new-password", "new-password-with-error"}, reqFactory, pwdRepo, configRepo)

	testassert.SliceContains(t, ui.PasswordPrompts, testassert.Lines{
		{"Current Password"},
		{"New Password"},
		{"Verify Password"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"Password verification does not match"},
	})

	assert.Equal(t, pwdRepo.UpdateNewPassword, "")
}

func TestWhenCurrentPasswordDoesNotMatch(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{ValidAccessTokenSuccess: true}
	pwdRepo := &testapi.FakePasswordRepo{UpdateUnauthorized: true}
	configRepo := &testconfig.FakeConfigRepository{}
	ui := callPassword([]string{"old-password", "new-password", "new-password"}, reqFactory, pwdRepo, configRepo)

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

	assert.Equal(t, pwdRepo.UpdateNewPassword, "new-password")
	assert.Equal(t, pwdRepo.UpdateOldPassword, "old-password")
}

func callPassword(inputs []string, reqFactory *testreq.FakeReqFactory, pwdRepo *testapi.FakePasswordRepo, configRepo *testconfig.FakeConfigRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{Inputs: inputs}

	ctxt := testcmd.NewContext("passwd", []string{})
	cmd := NewPassword(ui, pwdRepo, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
