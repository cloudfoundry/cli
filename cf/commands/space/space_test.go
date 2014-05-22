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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package space_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/space"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

func callShowSpace(args []string, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	config := testconfig.NewRepositoryWithDefaults()
	cmd := NewShowSpace(ui, config)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestShowSpaceRequirements", func() {
		args := []string{"my-space"}

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callShowSpace(args, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callShowSpace(args, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callShowSpace(args, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})

	It("TestShowSpaceInfoSuccess", func() {
		org := models.OrganizationFields{}
		org.Name = "my-org"

		app := models.ApplicationFields{}
		app.Name = "app1"
		app.Guid = "app1-guid"
		apps := []models.ApplicationFields{app}

		domain := models.DomainFields{}
		domain.Name = "domain1"
		domain.Guid = "domain1-guid"
		domains := []models.DomainFields{domain}

		serviceInstance := models.ServiceInstanceFields{}
		serviceInstance.Name = "service1"
		serviceInstance.Guid = "service1-guid"
		services := []models.ServiceInstanceFields{serviceInstance}

		space := models.Space{}
		space.Name = "space1"
		space.Organization = org
		space.Applications = apps
		space.Domains = domains
		space.ServiceInstances = services

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: space}
		ui := callShowSpace([]string{"space1"}, requirementsFactory)
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting info for space", "space1", "my-org", "my-user"},
			[]string{"OK"},
			[]string{"space1"},
			[]string{"Org", "my-org"},
			[]string{"Apps", "app1"},
			[]string{"Domains", "domain1"},
			[]string{"Services", "service1"},
		))
	})
})
