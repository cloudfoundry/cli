package securitygroup_test

import (
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"

	testapi "github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("staging-security-groups command", func() {
	var (
		ui                           *testterm.FakeUI
		configRepo                   configuration.ReadWriter
		fakeStagingSecurityGroupRepo *testapi.FakeStagingSecurityGroupsRepo
		cmd                          command.Command
		requirementsFactory          *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		fakeStagingSecurityGroupRepo = &testapi.FakeStagingSecurityGroupsRepo{}
		cmd = NewListStagingSecurityGroups(ui, configRepo, fakeStagingSecurityGroupRepo)
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) testcmd.RunCommandResult {
		return testcmd.RunCommandMoreBetter(cmd, requirementsFactory, args...)
	}

	Describe("requirements", func() {
		It("should fail when not logged in", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when there are some security groups set for staging", func() {
			BeforeEach(func() {
				fakeStagingSecurityGroupRepo.ListReturns([]models.SecurityGroupFields{
					{Name: "hiphopopotamus"},
					{Name: "my lyrics are bottomless"},
					{Name: "steve"},
				}, nil)
			})

			It("shows the user the name of the security groups for staging", func() {
				Expect(runCommand()).To(HaveSucceeded())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Acquiring", "staging security group", "my-user"},
					[]string{"hiphopopotamus"},
					[]string{"my lyrics are bottomless"},
					[]string{"steve"},
				))
			})
		})

		Context("when the API returns an error", func() {
			BeforeEach(func() {
				fakeStagingSecurityGroupRepo.ListReturns(nil, errors.New("uh oh"))
			})

			It("fails loudly", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when there are no security groups set for staging", func() {
			It("tells the user that there are none", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"No", "staging security group", "set"},
				))
			})
		})
	})
})
