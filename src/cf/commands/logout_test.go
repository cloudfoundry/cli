package commands_test

import (
	"cf/commands"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestLogoutClearsAccessTokenOrgAndSpace", func() {

			org := models.OrganizationFields{}
			org.Name = "MyOrg"

			space := models.SpaceFields{}
			space.Name = "MySpace"

			config := &configuration.Configuration{}
			config.AccessToken = "MyAccessToken"
			config.OrganizationFields = org
			config.SpaceFields = space

			ui := new(testterm.FakeUI)

			l := commands.NewLogout(ui, config)
			l.Run(nil)

			assert.Empty(mr.T(), config.AccessToken)
			assert.Equal(mr.T(), config.OrganizationFields, models.OrganizationFields{})
			assert.Equal(mr.T(), config.SpaceFields, models.SpaceFields{})
		})
	})
}
