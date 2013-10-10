package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
	"testing"
)

func TestLogoutClearsAccessTokenOrgAndSpace(t *testing.T) {
	configRepo := &testconfig.FakeConfigRepository{}
	config, _ := configRepo.Get()
	config.AccessToken = "MyAccessToken"
	config.Organization = cf.Organization{Name: "MyOrg"}
	config.Space = cf.Space{Name: "MySpace"}

	ui := new(testterm.FakeUI)

	l := NewLogout(ui, configRepo)
	l.Run(nil)

	updatedConfig, err := configRepo.Get()
	assert.NoError(t, err)

	assert.Empty(t, updatedConfig.AccessToken)
	assert.Equal(t, updatedConfig.Organization, cf.Organization{})
	assert.Equal(t, updatedConfig.Space, cf.Space{})
}
