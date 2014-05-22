package domain_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/domain"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
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

var _ = Describe("delete-domain command", func() {
	var (
		cmd                 *DeleteDomain
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		domainRepo          *testapi.FakeDomainRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"yes"},
		}

		domainRepo = &testapi.FakeDomainRepository{}
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:       true,
			TargetedOrgSuccess: true,
		}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) {
		cmd = NewDeleteDomain(ui, configRepo, domainRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("foo.com")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when the an org is not targetted", func() {
			requirementsFactory.TargetedOrgSuccess = false
			runCommand("foo.com")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the domain exists", func() {
		BeforeEach(func() {
			domainRepo.FindByNameInOrgDomain = models.DomainFields{
				Name: "foo.com",
				Guid: "foo-guid",
			}
		})

		It("deletes domains", func() {
			runCommand("foo.com")

			Expect(domainRepo.DeleteDomainGuid).To(Equal("foo-guid"))

			Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the domain foo.com"}))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting domain", "foo.com", "my-user"},
				[]string{"OK"},
			))
		})

		Context("when there is an error deleting the domain", func() {
			BeforeEach(func() {
				domainRepo.DeleteApiResponse = errors.New("failed badly")
			})

			It("show the error the user", func() {
				runCommand("foo.com")

				Expect(domainRepo.DeleteDomainGuid).To(Equal("foo-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting domain", "foo.com"},
					[]string{"FAILED"},
					[]string{"foo.com"},
					[]string{"failed badly"},
				))
			})
		})

		Context("when the user does not confirm", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"no"}
			})

			It("does nothing", func() {
				runCommand("foo.com")

				Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

				Expect(ui.Prompts).To(ContainSubstrings([]string{"delete", "foo.com"}))

				Expect(ui.Outputs).To(BeEmpty())
			})
		})

		Context("when the user provides the -f flag", func() {
			BeforeEach(func() {
				ui.Inputs = []string{}
			})

			It("skips confirmation", func() {
				runCommand("-f", "foo.com")

				Expect(domainRepo.DeleteDomainGuid).To(Equal("foo-guid"))
				Expect(ui.Prompts).To(BeEmpty())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting domain", "foo.com"},
					[]string{"OK"},
				))
			})
		})
	})

	Context("when a domain with the given name doesn't exist", func() {
		BeforeEach(func() {
			domainRepo.FindByNameInOrgApiResponse = errors.NewModelNotFoundError("Domain", "foo.com")
		})

		It("fails", func() {
			runCommand("foo.com")

			Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"OK"},
				[]string{"foo.com", "not found"},
			))
		})
	})

	Context("when there is an error finding the domain", func() {
		BeforeEach(func() {
			domainRepo.FindByNameInOrgApiResponse = errors.New("failed badly")
		})

		It("shows the error to the user", func() {
			runCommand("foo.com")

			Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"foo.com"},
				[]string{"failed badly"},
			))
		})
	})
})
