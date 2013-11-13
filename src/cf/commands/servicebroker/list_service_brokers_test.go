package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
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
		cf.ServiceBroker{
			Name: "service-broker-to-list-c",
			Guid: "service-broker-to-list-guid-c",
			Url:  "http://service-c-url.com",
		},
	}

	repo := &testapi.FakeServiceBrokerRepo{
		ServiceBrokers: serviceBrokers,
	}

	ui := callListServiceBrokers(t, []string{}, repo)

	assert.Contains(t, ui.Outputs[0], "Getting service brokers as")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Contains(t, ui.Outputs[1], "name")
	assert.Contains(t, ui.Outputs[1], "url")

	assert.Contains(t, ui.Outputs[2], "service-broker-to-list-a")
	assert.Contains(t, ui.Outputs[2], "http://service-a-url.com")

	assert.Contains(t, ui.Outputs[3], "service-broker-to-list-b")
	assert.Contains(t, ui.Outputs[3], "http://service-b-url.com")

	assert.Contains(t, ui.Outputs[4], "service-broker-to-list-c")
	assert.Contains(t, ui.Outputs[4], "http://service-c-url.com")
}

func TestListingServiceBrokersWhenNoneExist(t *testing.T) {
	repo := &testapi.FakeServiceBrokerRepo{
		ServiceBrokers: []cf.ServiceBroker{},
	}

	ui := callListServiceBrokers(t, []string{}, repo)

	assert.Contains(t, ui.Outputs[0], "Getting service brokers as")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "No service brokers found")
}

func TestListingServiceBrokersWhenFindFails(t *testing.T) {
	repo := &testapi.FakeServiceBrokerRepo{ListErr: true}

	ui := callListServiceBrokers(t, []string{}, repo)

	assert.Contains(t, ui.Outputs[0], "Getting service brokers as")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "FAILED")
}

func callListServiceBrokers(t *testing.T, args []string, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	ctxt := testcmd.NewContext("service-brokers", args)
	cmd := NewListServiceBrokers(ui, config, serviceBrokerRepo)
	testcmd.RunCommand(cmd, ctxt, &testreq.FakeReqFactory{})

	return
}
