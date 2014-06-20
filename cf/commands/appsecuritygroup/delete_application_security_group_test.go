package appsecuritygroup_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/appsecuritygroup"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-application-security-group command", func() {
	var (
		ui                   *testterm.FakeUI
		appSecurityGroupRepo *testapi.FakeAppSecurityGroup
		requirementsFactory  *testreq.FakeReqFactory
		configRepo           configuration.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		appSecurityGroupRepo = &testapi.FakeAppSecurityGroup{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) {
		cmd := NewDeleteAppSecurityGroup(ui, configRepo, appSecurityGroupRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("should fail if not logged in", func() {
			runCommand("my-group")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("should fail with usage when not provided a single argument", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("whoops", "I can't believe", "I accidentally", "the whole thing")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when the group with the given name exists", func() {
			BeforeEach(func() {
				appSecurityGroupRepo.ReadReturns.Fields = models.ApplicationSecurityGroupFields{
					Name: "my-group",
					Guid: "group-guid",
				}
			})

			It("should delete the application security group", func() {
				runCommand("my-group")
				Expect(appSecurityGroupRepo.ReadCalledWith.Name).To(Equal("my-group"))
				Expect(appSecurityGroupRepo.DeleteCalledWith.Guid).To(Equal("group-guid"))
			})

			It("tells the user what it's about to do", func() {
				runCommand("my-group")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "application security group", "my-group", "my-user"},
					[]string{"OK"},
				))
			})
		})

		It("fails and warns the user if a group with that name could not be found", func() {
			appSecurityGroupRepo.ReadReturns.Error = errors.New("pbbbbbbbbbbt")
			runCommand("whoops")

			Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
		})

		It("fails and warns the user if deleting fails", func() {
			appSecurityGroupRepo.DeleteReturns.Error = errors.New("raspberry")
			runCommand("whoops")

			Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
		})
	})
})
