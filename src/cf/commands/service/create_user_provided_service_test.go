package service_test

import (
	"cf/api"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateUserProvidedService(t *testing.T) {
	serviceRepo := &testhelpers.FakeServiceRepo{}
	fakeUI := callCreateUserProvidedService(
		[]string{"my-custom-service", `"foo, bar, baz"`},
		[]string{"foo value", "bar value", "baz value"},
		serviceRepo,
	)

	assert.Contains(t, fakeUI.Prompts[0], "foo")
	assert.Contains(t, fakeUI.Prompts[1], "bar")
	assert.Contains(t, fakeUI.Prompts[2], "baz")

	assert.Equal(t, serviceRepo.CreateUserProvidedServiceInstanceName, "my-custom-service")
	assert.Equal(t, serviceRepo.CreateUserProvidedServiceInstanceParameters, map[string]string{
		"foo": "foo value",
		"bar": "bar value",
		"baz": "baz value",
	})

	assert.Contains(t, fakeUI.Outputs[0], "Creating user provided service")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateUserProvidedServiceWithNoParameterList(t *testing.T) {
	serviceRepo := &testhelpers.FakeServiceRepo{}
	fakeUI := callCreateUserProvidedService(
		[]string{"my-custom-service"},
		[]string{},
		serviceRepo,
	)

	assert.Contains(t, fakeUI.Outputs[0], "FAILED")
}

func callCreateUserProvidedService(args []string, inputs []string, serviceRepo api.ServiceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = &testhelpers.FakeUI{Inputs: inputs}
	ctxt := testhelpers.NewContext("create-user-provided-service", args)
	cmd := NewCreateUserProvidedService(fakeUI, serviceRepo)
	reqFactory := &testhelpers.FakeReqFactory{}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
