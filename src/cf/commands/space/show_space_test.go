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

package space_test

import (
	. "cf/commands/space"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callShowSpace(args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("space", args)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewShowSpace(ui, config)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestShowSpaceRequirements", func() {
		args := []string{"my-space"}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callShowSpace(args, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callShowSpace(args, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callShowSpace(args, reqFactory)
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

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true, Space: space}
		ui := callShowSpace([]string{"space1"}, reqFactory)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting info for space", "space1", "my-org", "my-user"},
			{"OK"},
			{"space1"},
			{"Org", "my-org"},
			{"Apps", "app1"},
			{"Domains", "domain1"},
			{"Services", "service1"},
		})
	})
})
