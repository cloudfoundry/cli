package organization_test

import (
	. "cf/commands/organization"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateOrgFailsWithUsage(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	fakeUI := callCreateOrg([]string{}, reqFactory, orgRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestCreateOrgRequirements(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}
	callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestCreateOrg(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	fakeUI := callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating org")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Equal(t, orgRepo.CreateName, "my-org")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateOrgWhenAlreadyExists(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{CreateOrgExists: true}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	fakeUI := callCreateOrg([]string{"my-org"}, reqFactory, orgRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating org")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-org")
	assert.Contains(t, fakeUI.Outputs[2], "already exists")
}

func callCreateOrg(args []string, reqFactory *testhelpers.FakeReqFactory, orgRepo *testhelpers.FakeOrgRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("create-org", args)
	cmd := NewCreateOrg(fakeUI, orgRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
