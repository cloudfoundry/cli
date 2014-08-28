package organization_test

import (
	"github.com/cloudfoundry/cli/cf/errors"

	test_org "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/organization"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-org command", func() {
	var (
		config              configuration.ReadWriter
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		orgRepo             *test_org.FakeOrganizationRepository
		org                 models.Organization
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"y"},
		}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}

		org = models.Organization{}
		org.Name = "org-to-delete"
		org.Guid = "org-to-delete-guid"
		orgRepo = &test_org.FakeOrganizationRepository{}

		orgRepo.ListOrgsReturns([]models.Organization{org}, nil)
		orgRepo.FindByNameReturns(org, nil)
	})

	runCommand := func(args ...string) {
		cmd := NewDeleteOrg(ui, config, orgRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	It("fails requirements when not logged in", func() {
		runCommand("some-org-name")

		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("fails with usage if no arguments are given", func() {
		runCommand()
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when deleting the currently targeted org", func() {
			It("untargets the deleted org", func() {
				config.SetOrganizationFields(org.OrganizationFields)
				runCommand("org-to-delete")

				Expect(config.OrganizationFields()).To(Equal(models.OrganizationFields{}))
				Expect(config.SpaceFields()).To(Equal(models.SpaceFields{}))
			})
		})

		Context("when deleting an org that is not targeted", func() {
			BeforeEach(func() {
				otherOrgFields := models.OrganizationFields{}
				otherOrgFields.Guid = "some-other-org-guid"
				otherOrgFields.Name = "some-other-org"
				config.SetOrganizationFields(otherOrgFields)

				spaceFields := models.SpaceFields{}
				spaceFields.Name = "some-other-space"
				config.SetSpaceFields(spaceFields)
			})

			It("deletes the org with the given name", func() {
				runCommand("org-to-delete")

				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the org org-to-delete"}))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "org-to-delete"},
					[]string{"OK"},
				))

				Expect(orgRepo.DeleteArgsForCall(0)).To(Equal("org-to-delete-guid"))
			})

			It("does not untarget the org and space", func() {
				runCommand("org-to-delete")

				Expect(config.OrganizationFields().Name).To(Equal("some-other-org"))
				Expect(config.SpaceFields().Name).To(Equal("some-other-space"))
			})
		})

		It("does not prompt when the -f flag is given", func() {
			ui.Inputs = []string{}
			runCommand("-f", "org-to-delete")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting", "org-to-delete"},
				[]string{"OK"},
			))

			Expect(orgRepo.DeleteArgsForCall(0)).To(Equal("org-to-delete-guid"))
		})

		It("warns the user when the org does not exist", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.NewModelNotFoundError("Organization", "org org-to-delete does not exist"))

			runCommand("org-to-delete")

			Expect(orgRepo.DeleteCallCount()).To(Equal(0))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting", "org-to-delete"},
				[]string{"OK"},
			))
			Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"org-to-delete", "does not exist."}))
		})
	})
})
