package commands_test

import (
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestPasswordRequiresLogin(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: false}
	callPassword([]string{}, reqFactory, &testhelpers.FakePasswordRepo{})
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true}
	callPassword([]string{"", "", ""}, reqFactory, &testhelpers.FakePasswordRepo{})
	assert.True(t, testhelpers.CommandDidPassRequirements)
}

func TestPasswordCanBeChanged(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	pwdRepo := &testhelpers.FakePasswordRepo{Score: "meh"}
	ui := callPassword([]string{"old-password", "new-password", "new-password"}, reqFactory, pwdRepo)

	assert.Contains(t, ui.Prompts[0], "Current Password")
	assert.Contains(t, ui.Prompts[1], "New Password")
	assert.Contains(t, ui.Prompts[2], "Verify Password")

	assert.Equal(t, pwdRepo.ScoredPassword, "new-password")
	assert.Contains(t, ui.Outputs[0], "Your password strength is: meh")

	assert.Contains(t, ui.Outputs[1], "Changing password...")
	assert.Equal(t, pwdRepo.UpdateNewPassword, "new-password")
	assert.Equal(t, pwdRepo.UpdateOldPassword, "old-password")
	assert.Contains(t, ui.Outputs[2], "OK")
}

func callPassword(inputs []string, reqFactory *testhelpers.FakeReqFactory, pwdRepo *testhelpers.FakePasswordRepo) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{Inputs: inputs}

	ctxt := testhelpers.NewContext("passwd", []string{})
	cmd := NewPassword(ui, pwdRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
