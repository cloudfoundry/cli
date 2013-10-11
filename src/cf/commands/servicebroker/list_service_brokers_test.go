package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testterm "testhelpers/terminal"
	"testing"
)

func TestListServiceBrokers(t *testing.T) {
	serviceBrokers := []cf.ServiceBroker{
		cf.ServiceBroker{
			Name: "service-broker-to-list-a",
			Guid: "service-broker-to-list-guid-a",
			Url:  "http://service-a-url.com",
		},
		cf.ServiceBroker{
			Name: "service-broker-to-list-b",
			Guid: "service-broker-to-list-guid-b",
			Url:  "http://service-b-url.com",
		},
	}

	repo := &testapi.FakeServiceBrokerRepo{
		FindAllServiceBrokers: serviceBrokers,
	}

	ui := &testterm.FakeUI{}

	cmd := NewListServiceBrokers(ui, repo)
	cmd.Run(testcmd.NewContext("service-brokers", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting service brokers...")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "Name")
	assert.Contains(t, ui.Outputs[3], "URL")

	assert.Contains(t, ui.Outputs[4], "service-broker-to-list-a")
	assert.Contains(t, ui.Outputs[4], "http://service-a-url.com")

	assert.Contains(t, ui.Outputs[5], "service-broker-to-list-b")
	assert.Contains(t, ui.Outputs[5], "http://service-b-url.com")
}

func TestListingServiceBrokersWhenNoneExist(t *testing.T) {
	repo := &testapi.FakeServiceBrokerRepo{
		FindAllServiceBrokers: []cf.ServiceBroker{},
	}

	ui := &testterm.FakeUI{}

	cmd := NewListServiceBrokers(ui, repo)
	cmd.Run(testcmd.NewContext("service-brokers", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting service brokers...")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "No service brokers found")
}

func TestListingServiceBrokersWhenFindFails(t *testing.T) {
	repo := &testapi.FakeServiceBrokerRepo{FindAllErr: true}
	ui := &testterm.FakeUI{}

	cmd := NewListServiceBrokers(ui, repo)
	cmd.Run(testcmd.NewContext("service-brokers", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting service brokers...")
	assert.Contains(t, ui.Outputs[1], "FAILED")
}
