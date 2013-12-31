package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteServiceBrokerFailsWithUsage(t *testing.T) {
	ui, _, _ := deleteServiceBroker(t, "y", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _, _ = deleteServiceBroker(t, "y", []string{"my-broker"})
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}

	reqFactory.LoginSuccess = false
	callDeleteServiceBroker(t, []string{"-f", "my-broker"}, reqFactory, repo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callDeleteServiceBroker(t, []string{"-f", "my-broker"}, reqFactory, repo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteConfirmingWithY(t *testing.T) {
	ui, _, repo := deleteServiceBroker(t, "y", []string{"service-broker-to-delete"})

	assert.Equal(t, repo.FindByNameName, "service-broker-to-delete")
	assert.Equal(t, repo.DeletedServiceBrokerGuid, "service-broker-to-delete-guid")
	assert.Equal(t, len(ui.Outputs), 2)
	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Really delete", "service-broker-to-delete"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service broker", "service-broker-to-delete", "my-user"},
		{"OK"},
	})
}

func TestDeleteConfirmingWithYes(t *testing.T) {
	ui, _, repo := deleteServiceBroker(t, "Yes", []string{"service-broker-to-delete"})

	assert.Equal(t, repo.FindByNameName, "service-broker-to-delete")
	assert.Equal(t, repo.DeletedServiceBrokerGuid, "service-broker-to-delete-guid")
	assert.Equal(t, len(ui.Outputs), 2)
	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Really delete", "service-broker-to-delete"},
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service broker", "service-broker-to-delete", "my-user"},
		{"OK"},
	})
}

func TestDeleteWithForceOption(t *testing.T) {
	serviceBroker := cf.ServiceBroker{}
	serviceBroker.Name = "service-broker-to-delete"
	serviceBroker.Guid = "service-broker-to-delete-guid"

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo := &testapi.FakeServiceBrokerRepo{FindByNameServiceBroker: serviceBroker}
	ui := callDeleteServiceBroker(t, []string{"-f", "service-broker-to-delete"}, reqFactory, repo)

	assert.Equal(t, repo.FindByNameName, "service-broker-to-delete")
	assert.Equal(t, repo.DeletedServiceBrokerGuid, "service-broker-to-delete-guid")
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Equal(t, len(ui.Outputs), 2)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service broker", "service-broker-to-delete", "my-user"},
		{"OK"},
	})
}

func TestDeleteAppThatDoesNotExist(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo := &testapi.FakeServiceBrokerRepo{FindByNameNotFound: true}
	ui := callDeleteServiceBroker(t, []string{"-f", "service-broker-to-delete"}, reqFactory, repo)

	assert.Equal(t, repo.FindByNameName, "service-broker-to-delete")
	assert.Equal(t, repo.DeletedServiceBrokerGuid, "")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service broker", "service-broker-to-delete"},
		{"OK"},
		{"service-broker-to-delete", "does not exist"},
	})
}

func callDeleteServiceBroker(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete-service-broker", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewDeleteServiceBroker(ui, config, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func deleteServiceBroker(t *testing.T, confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) {
	serviceBroker := cf.ServiceBroker{}
	serviceBroker.Name = "service-broker-to-delete"
	serviceBroker.Guid = "service-broker-to-delete-guid"

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	repo = &testapi.FakeServiceBrokerRepo{FindByNameServiceBroker: serviceBroker}
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space2 := cf.SpaceFields{}
	space2.Name = "my-space"
	org2 := cf.OrganizationFields{}
	org2.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space2,
		OrganizationFields: org2,
		AccessToken:        token,
	}

	ctxt := testcmd.NewContext("delete-service-broker", args)
	cmd := NewDeleteServiceBroker(ui, config, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
