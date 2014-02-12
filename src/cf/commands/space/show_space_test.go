package space_test

import (
	. "cf/commands/space"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callShowSpace(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
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
		callShowSpace(mr.T(), args, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callShowSpace(mr.T(), args, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callShowSpace(mr.T(), args, reqFactory)
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
		ui := callShowSpace(mr.T(), []string{"space1"}, reqFactory)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
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
