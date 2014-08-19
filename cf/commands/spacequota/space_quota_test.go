package spacequota_test

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/spacequota"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("quotas command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *fakes.FakeSpaceQuotaRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = &fakes.FakeSpaceQuotaRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	runCommand := func(args ...string) bool {
		cmd := NewSpaceQuota(ui, testconfig.NewRepositoryWithDefaults(), quotaRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand("foo")).ToNot(HavePassedRequirements())
		})

		It("requires the user to target an org", func() {
			requirementsFactory.TargetedOrgSuccess = false
			Expect(runCommand("bar")).ToNot(HavePassedRequirements())
		})

		It("fails when a quota name is not provided", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Context("when logged in", func() {
		JustBeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			Expect(runCommand("quota-name")).To(HavePassedRequirements())
		})

		Context("when quotas exist", func() {
			BeforeEach(func() {
				quotaRepo.FindByNameReturns(
					models.SpaceQuota{
						Name:                    "quota-name",
						MemoryLimit:             1024,
						InstanceMemoryLimit:     512,
						RoutesLimit:             111,
						ServicesLimit:           222,
						NonBasicServicesAllowed: true,
						OrgGuid:                 "my-org-guid",
					}, nil)
			})

			It("lists the specific quota info", func() {
				Expect(quotaRepo.FindByNameArgsForCall(0)).To(Equal("quota-name"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting space quota quota-name info as", "my-user"},
					[]string{"OK"},
					[]string{"total memory limit", "1G"},
					[]string{"instance memory limit", "512M"},
					[]string{"routes", "111"},
					[]string{"service", "222"},
					[]string{"non basic services", "allowed"},
				))
			})
		})

		Context("when an error occurs fetching quotas", func() {
			BeforeEach(func() {
				quotaRepo.FindByNameReturns(models.SpaceQuota{}, errors.New("I haz a borken!"))
			})

			It("prints an error", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting space quota quota-name info as", "my-user"},
					[]string{"FAILED"},
				))
			})
		})
	})

})
