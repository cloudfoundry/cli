package domain_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/domain"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-shared-domain command", func() {
	var (
		ui                  *testterm.FakeUI
		domainRepo          *testapi.FakeDomainRepository
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		domainRepo = &testapi.FakeDomainRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewDeleteSharedDomain(ui, configRepo, domainRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails if you are not logged in", func() {
			requirementsFactory.TargetedOrgSuccess = true

			Expect(runCommand("foo.com")).To(BeFalse())
		})

		It("fails if an organiztion is not targeted", func() {
			requirementsFactory.LoginSuccess = true

			Expect(runCommand("foo.com")).To(BeFalse())
		})
	})

	Context("when logged in and targeted an organiztion", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			domainRepo.FindByNameInOrgDomain = []models.DomainFields{
				models.DomainFields{
					Name:   "foo.com",
					Guid:   "foo-guid",
					Shared: true,
				},
			}
		})

		Describe("and the command is invoked interactively", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"y"}
			})

			It("when the domain is not found it tells the user", func() {
				domainRepo.FindByNameInOrgApiResponse = errors.NewModelNotFoundError("Domain", "foo.com")
				runCommand("foo.com")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting domain", "foo.com"},
					[]string{"OK"},
					[]string{"foo.com", "not found"},
				))
			})

			It("fails when the api returns an error", func() {
				domainRepo.FindByNameInOrgApiResponse = errors.New("couldn't find the droids you're lookin for")
				runCommand("foo.com")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting domain", "foo.com"},
					[]string{"FAILED"},
					[]string{"foo.com"},
					[]string{"couldn't find the droids you're lookin for"},
				))
			})

			It("fails when deleting the domain encounters an error", func() {
				domainRepo.DeleteSharedApiResponse = errors.New("failed badly")
				runCommand("foo.com")

				Expect(domainRepo.DeleteSharedDomainGuid).To(Equal("foo-guid"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting domain", "foo.com"},
					[]string{"FAILED"},
					[]string{"foo.com"},
					[]string{"failed badly"},
				))
			})

			It("Prompts a user to delete the shared domain", func() {
				runCommand("foo.com")

				Expect(domainRepo.DeleteSharedDomainGuid).To(Equal("foo-guid"))
				Expect(ui.Prompts).To(ContainSubstrings([]string{"delete", "domain", "foo.com"}))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting domain", "foo.com"},
					[]string{"OK"},
				))
			})
		})

		It("skips confirmation if the force flag is passed", func() {
			runCommand("-f", "foo.com")

			Expect(domainRepo.DeleteSharedDomainGuid).To(Equal("foo-guid"))
			Expect(ui.Prompts).To(BeEmpty())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting domain", "foo.com"},
				[]string{"OK"},
			))
		})
	})
})
