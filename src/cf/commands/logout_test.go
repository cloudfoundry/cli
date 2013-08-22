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
	config, err := configuration.Load()
	assert.NoError(t, err)

	config.AccessToken = "MyAccessToken"
	config.Organization = cf.Organization{Name: "MyOrg"}
	config.Space = cf.Space{Name: "MyOrg"}

	err = config.Save()
	assert.NoError(t, err)

	ui := new(testhelpers.FakeUI)

	l := NewLogout(ui)
	l.Run(nil)

	config, err = configuration.Load()
	assert.NoError(t, err)

	assert.Empty(t, config.AccessToken)
	assert.Equal(t, config.Organization, cf.Organization{})
	assert.Equal(t, config.Space, cf.Space{})
}
