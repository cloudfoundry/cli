/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package organization_test

import (
	"cf/commands/organization"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("set-quota command", func() {
	var reqFactory *testreq.FakeReqFactory
	var quotaRepo *testapi.FakeQuotaRepository

	BeforeEach(func() {
		reqFactory = &testreq.FakeReqFactory{}
		quotaRepo = &testapi.FakeQuotaRepository{}
	})

	It("fails with usage when provided too many or two few args", func() {
		ui := callSetQuota([]string{}, reqFactory, quotaRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetQuota([]string{"org"}, reqFactory, quotaRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callSetQuota([]string{"org", "quota", "extra-stuff"}, reqFactory, quotaRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails requirements when not logged in", func() {
		callSetQuota([]string{"my-org", "my-quota"}, reqFactory, quotaRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			reqFactory.LoginSuccess = true
		})

		It("passes requirements when provided two args", func() {
			callSetQuota([]string{"my-org", "my-quota"}, reqFactory, quotaRepo)
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
			Expect(reqFactory.OrganizationName).To(Equal("my-org"))
		})

		It("TestSetQuota", func() {
			org := models.Organization{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"

			quota := models.QuotaFields{}
			quota.Name = "my-found-quota"
			quota.Guid = "my-quota-guid"

			quotaRepo.FindByNameQuota = quota
			reqFactory.Organization = org

			ui := callSetQuota([]string{"my-org", "my-quota"}, reqFactory, quotaRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Setting quota", "my-found-quota", "my-org", "my-user"},
				{"OK"},
			})

			Expect(quotaRepo.FindByNameName).To(Equal("my-quota"))
			Expect(quotaRepo.UpdateOrgGuid).To(Equal("my-org-guid"))
			Expect(quotaRepo.UpdateQuotaGuid).To(Equal("my-quota-guid"))
		})
	})
})

func callSetQuota(args []string, reqFactory *testreq.FakeReqFactory, quotaRepo *testapi.FakeQuotaRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-quota", args)

	token := configuration.TokenInfo{
		Username: "my-user",
	}

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"

	configRepo := testconfig.NewRepositoryWithAccessToken(token)
	configRepo.SetSpaceFields(spaceFields)
	configRepo.SetOrganizationFields(orgFields)

	cmd := organization.NewSetQuota(ui, configRepo, quotaRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
