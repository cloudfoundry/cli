package commands_test

import (
	"cf/commands"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

var _ = Describe("logout command", func() {
	var config configuration.Repository
	BeforeEach(func() {
		org := models.OrganizationFields{}
		org.Name = "MyOrg"

		space := models.SpaceFields{}
		space.Name = "MySpace"

		config = testconfig.NewRepository()
		config.SetAccessToken("MyAccessToken")
		config.SetOrganizationFields(org)
		config.SetSpaceFields(space)
		ui := new(testterm.FakeUI)

		l := commands.NewLogout(ui, config)
		l.Run(nil)
	})

	It("clears access token from the config", func() {
		Expect(config.AccessToken()).To(Equal(""))
	})

	It("clears organization fields from the config", func() {
		Expect(config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
	})

	It("clears space fields from the config", func() {
		Expect(config.SpaceFields()).To(Equal(models.SpaceFields{}))
	})
})
