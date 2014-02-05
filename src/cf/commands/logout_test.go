package commands_test

import (
	"cf"
	"cf/commands"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestLogoutClearsAccessTokenOrgAndSpace", func() {

			org := cf.OrganizationFields{}
			org.Name = "MyOrg"

			space := cf.SpaceFields{}
			space.Name = "MySpace"

			config := &configuration.Configuration{}
			config.AccessToken = "MyAccessToken"
			config.OrganizationFields = org
			config.SpaceFields = space

			ui := new(testterm.FakeUI)

			l := commands.NewLogout(ui, config)
			l.Run(nil)

			assert.Empty(mr.T(), config.AccessToken)
			assert.Equal(mr.T(), config.OrganizationFields, cf.OrganizationFields{})
			assert.Equal(mr.T(), config.SpaceFields, cf.SpaceFields{})
		})
	})
}
