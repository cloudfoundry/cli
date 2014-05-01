package servicebroker_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/servicebroker"
	"github.com/cloudfoundry/cli/cf/configuration"
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

func callListServiceBrokers(args []string, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	config := testconfig.NewRepositoryWithDefaults()
	ctxt := testcmd.NewContext("service-brokers", args)
	cmd := NewListServiceBrokers(ui, config, serviceBrokerRepo)
	testcmd.RunCommand(cmd, ctxt, &testreq.FakeReqFactory{})

	return
}

var _ = Describe("service-brokers command", func() {
	var (
		ui                  *testterm.FakeUI
		config              configuration.Repository
		cmd                 ListServiceBrokers
		repo                *testapi.FakeServiceBrokerRepo
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		repo = &testapi.FakeServiceBrokerRepo{}
		cmd = NewListServiceBrokers(ui, config, repo)
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			ctxt := testcmd.NewContext("service-brokers", []string{})
			testcmd.RunCommand(cmd, ctxt, requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("lists service brokers", func() {
		repo.ServiceBrokers = []models.ServiceBroker{models.ServiceBroker{
			Name: "service-broker-to-list-a",
			Guid: "service-broker-to-list-guid-a",
			Url:  "http://service-a-url.com",
		}, models.ServiceBroker{
			Name: "service-broker-to-list-b",
			Guid: "service-broker-to-list-guid-b",
			Url:  "http://service-b-url.com",
		}, models.ServiceBroker{
			Name: "service-broker-to-list-c",
			Guid: "service-broker-to-list-guid-c",
			Url:  "http://service-c-url.com",
		}}

		context := testcmd.NewContext("service-brokers", []string{})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting service brokers as", "my-user"},
			[]string{"name", "url"},
			[]string{"service-broker-to-list-a", "http://service-a-url.com"},
			[]string{"service-broker-to-list-b", "http://service-b-url.com"},
			[]string{"service-broker-to-list-c", "http://service-c-url.com"},
		))
	})

	It("says when no service brokers were found", func() {
		context := testcmd.NewContext("service-brokers", []string{})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting service brokers as", "my-user"},
			[]string{"No service brokers found"},
		))
	})

	It("reports errors when listing service brokers", func() {
		repo.ListErr = true
		context := testcmd.NewContext("service-brokers", []string{})
		testcmd.RunCommand(cmd, context, requirementsFactory)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting service brokers as ", "my-user"},
			[]string{"FAILED"},
		))
	})
})
