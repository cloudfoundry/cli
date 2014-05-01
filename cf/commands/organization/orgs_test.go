package organization_test

import (
	"github.com/cloudfoundry/cli/cf/commands/organization"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("org command", func() {
	var (
		ui                  *testterm.FakeUI
		orgRepo             *testapi.FakeOrgRepository
		configRepo          configuration.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
	)

	runCommand := func() {
		cmd := organization.NewListOrgs(ui, configRepo, orgRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("orgs", []string{}), requirementsFactory)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		orgRepo = &testapi.FakeOrgRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when there are orgs to be listed", func() {
		BeforeEach(func() {
			org1 := models.Organization{}
			org1.Name = "Organization-1"

			org2 := models.Organization{}
			org2.Name = "Organization-2"

			org3 := models.Organization{}
			org3.Name = "Organization-3"

			orgRepo.Organizations = []models.Organization{org1, org2, org3}
		})

		It("lists orgs", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting orgs as my-user"},
				[]string{"Organization-1"},
				[]string{"Organization-2"},
				[]string{"Organization-3"},
			))
		})
	})

	It("tells the user when no orgs were found", func() {
		runCommand()

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting orgs as my-user"},
			[]string{"No orgs found"},
		))
	})
})
