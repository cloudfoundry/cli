package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestLogoutClearsAccessTokenOrgAndSpace(t *testing.T) {
	config := &configuration.Configuration{}
	config.AccessToken = "MyAccessToken"
	config.Organization = cf.Organization{Name: "MyOrg"}
	config.Space = cf.Space{Name: "MyOrg"}

	ui := new(testhelpers.FakeUI)

	l := NewLogout(ui, config)
	l.Run(nil)

	updatedConfig, err := configuration.Get()
	assert.NoError(t, err)

	assert.Empty(t, updatedConfig.AccessToken)
	assert.Equal(t, updatedConfig.Organization, cf.Organization{})
	assert.Equal(t, updatedConfig.Space, cf.Space{})
}
