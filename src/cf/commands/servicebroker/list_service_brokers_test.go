package servicebroker_test

import (
	. "cf/commands/servicebroker"
	"cf/models"
	. "github.com/onsi/ginkgo"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callListServiceBrokers(args []string, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	config := testconfig.NewRepositoryWithDefaults()
	ctxt := testcmd.NewContext("service-brokers", args)
	cmd := NewListServiceBrokers(ui, config, serviceBrokerRepo)
	testcmd.RunCommand(cmd, ctxt, &testreq.FakeReqFactory{})

	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestListServiceBrokers", func() {
		broker := models.ServiceBroker{}
		broker.Name = "service-broker-to-list-a"
		broker.Guid = "service-broker-to-list-guid-a"
		broker.Url = "http://service-a-url.com"
		broker2 := models.ServiceBroker{}
		broker2.Name = "service-broker-to-list-b"
		broker2.Guid = "service-broker-to-list-guid-b"
		broker2.Url = "http://service-b-url.com"
		broker3 := models.ServiceBroker{}
		broker3.Name = "service-broker-to-list-c"
		broker3.Guid = "service-broker-to-list-guid-c"
		broker3.Url = "http://service-c-url.com"
		serviceBrokers := []models.ServiceBroker{broker, broker2, broker3}

		repo := &testapi.FakeServiceBrokerRepo{
			ServiceBrokers: serviceBrokers,
		}

		ui := callListServiceBrokers([]string{}, repo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting service brokers as", "my-user"},
			{"name", "url"},
			{"service-broker-to-list-a", "http://service-a-url.com"},
			{"service-broker-to-list-b", "http://service-b-url.com"},
			{"service-broker-to-list-c", "http://service-c-url.com"},
		})
	})
	It("TestListingServiceBrokersWhenNoneExist", func() {

		repo := &testapi.FakeServiceBrokerRepo{
			ServiceBrokers: []models.ServiceBroker{},
		}

		ui := callListServiceBrokers([]string{}, repo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting service brokers as", "my-user"},
			{"No service brokers found"},
		})
	})
	It("TestListingServiceBrokersWhenFindFails", func() {

		repo := &testapi.FakeServiceBrokerRepo{ListErr: true}

		ui := callListServiceBrokers([]string{}, repo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting service brokers as ", "my-user"},
			{"FAILED"},
		})
	})
})
