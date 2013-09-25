package commands_test

import (
	/*	"cf"*/
	/*	"cf/api"*/
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateOrgFailsWithUsage(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	fakeUI := callCreateOrganization([]string{}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callCreateOrganization([]string{"my-org"}, reqFactory, orgRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestCreateOrgRequirements(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callCreateOrganization([]string{"my-org"}, reqFactory, orgRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}
	callCreateOrganization([]string{"my-org"}, reqFactory, orgRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestCreateOrganization(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	fakeUI := callCreateOrganization([]string{"my-org"}, reqFactory, orgRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating organization")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Equal(t, orgRepo.CreateName, "my-org")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateOrganizationWhenAlreadyExists(t *testing.T){
	orgRepo := &testhelpers.FakeOrgRepository{CreateOrgExists: true}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	fakeUI := callCreateOrganization([]string{"my-org"}, reqFactory, orgRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating organization")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-org")
	assert.Contains(t, fakeUI.Outputs[2], "already exists.")
}

func callCreateOrganization(args []string, reqFactory *testhelpers.FakeReqFactory, orgRepo *testhelpers.FakeOrgRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("create-org", args)
	cmd := NewCreateOrganization(fakeUI, orgRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
