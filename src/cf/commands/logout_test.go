package commands_test

import (
	"cf"
	"cf/commands"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestLogoutClearsAccessTokenOrgAndSpace", func() {

			org := cf.OrganizationFields{}
			org.Name = "MyOrg"

			space := cf.SpaceFields{}
			space.Name = "MySpace"

			configRepo := &testconfig.FakeConfigRepository{}
			config, _ := configRepo.Get()
			config.AccessToken = "MyAccessToken"
			config.OrganizationFields = org
			config.SpaceFields = space

			ui := new(testterm.FakeUI)

			l := commands.NewLogout(ui, configRepo)
			l.Run(nil)

			updatedConfig, err := configRepo.Get()
			assert.NoError(mr.T(), err)

			assert.Empty(mr.T(), updatedConfig.AccessToken)
			assert.Equal(mr.T(), updatedConfig.OrganizationFields, cf.OrganizationFields{})
			assert.Equal(mr.T(), updatedConfig.SpaceFields, cf.SpaceFields{})
		})
	})
}
