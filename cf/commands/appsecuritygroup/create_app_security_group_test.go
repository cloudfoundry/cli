package appsecuritygroup_test

import (
	"github.com/cloudfoundry/cli/cf/errors"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/appsecuritygroup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-app-security-group", func() {
	var (
		ui                   *testterm.FakeUI
		appSecurityGroupRepo *testapi.FakeAppSecurityGroup
		requirementsFactory  *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		appSecurityGroupRepo = &testapi.FakeAppSecurityGroup{}
	})

	runCommand := func(args ...string) {
		cmd := NewCreateAppSecurityGroup(ui, appSecurityGroupRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			runCommand("the-security-group")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("succeeds when the user is logged in", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("the-security-group")
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails with usage when a name is not provided", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("creates the application security group", func() {
			runCommand("my-group")
			Expect(appSecurityGroupRepo.CreateArgsForCall(0).Name).To(Equal("my-group"))
		})

		It("displays message describing what its going to do", func() {
			runCommand("my-group")
			Expect(ui.Outputs).To(matchers.ContainSubstrings([]string{"Creating application security group my-group"}))
		})

		It("returns OK after successfully completing the command", func() {
			runCommand("my-group")
			Expect(ui.Outputs).To(matchers.ContainSubstrings(
				[]string{"Creating application security group my-group"},
				[]string{"OK"},
			))
		})

		It("allows the user to specify rules", func() {
			runCommand(
				"-rules",
				"[{\"protocol\":\"udp\",\"port\":\"8080-9090\",\"destination\":\"198.41.191.47/1\"}]",
				"app-security-groups-rule-everything-around-me",
			)

			Expect(appSecurityGroupRepo.CreateArgsForCall(0).Rules).To(Equal(
				"[{\"protocol\":\"udp\",\"port\":\"8080-9090\",\"destination\":\"198.41.191.47/1\"}]",
			))
		})

		Context("when creating a security group returns an error", func() {
			It("alerts the user when creating the security group fails", func() {
				appSecurityGroupRepo.CreateReturns(errors.New("Wops I failed"))
				runCommand("my-group")

				Expect(ui.Outputs).To(matchers.ContainSubstrings(
					[]string{"reating application security group", "my-group"},
					[]string{"FAILED"},
				))
			})
		})
	})
})
