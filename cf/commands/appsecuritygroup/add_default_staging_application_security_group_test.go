package appsecuritygroup_test

import (
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/appsecuritygroup"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("add-default-staging-application-security-group command", func() {
	var (
		ui                       *testterm.FakeUI
		requirementsFactory      *testreq.FakeReqFactory
		securityGroupRepo        *testapi.FakeAppSecurityGroup
		stagingSecurityGroupRepo *testapi.FakeStagingSecurityGroupsRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		securityGroupRepo = &testapi.FakeAppSecurityGroup{}
		stagingSecurityGroupRepo = &testapi.FakeStagingSecurityGroupsRepo{}
	})

	runCommand := func(args ...string) {
		cmd := NewAddToDefaultStagingGroup(ui, securityGroupRepo, stagingSecurityGroupRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			runCommand("name")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when a name is not provided", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the user is logged in and provides the name of a group", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			securityGroupRepo.ReadReturns.Fields = models.ApplicationSecurityGroupFields{Guid: "just-pretend-this-is-a-guid"}
		})

		It("adds the group to the default staging group set", func() {
			runCommand("a-security-group-name")

			Expect(securityGroupRepo.ReadCalledWith.Name).To(Equal("a-security-group-name"))
			Expect(stagingSecurityGroupRepo.AddToDefaultStagingSetArgsForCall(0).Guid).To(Equal("just-pretend-this-is-a-guid"))
		})
	})
})
