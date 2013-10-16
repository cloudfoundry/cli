package service_test

import (
	"cf/api"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateUserProvidedServiceWithParameterList(t *testing.T) {
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	fakeUI := callCreateUserProvidedService(
		[]string{"my-custom-service", `"foo, bar, baz"`},
		[]string{"foo value", "bar value", "baz value"},
		userProvidedServiceInstanceRepo,
	)

	assert.Contains(t, fakeUI.Prompts[0], "foo")
	assert.Contains(t, fakeUI.Prompts[1], "bar")
	assert.Contains(t, fakeUI.Prompts[2], "baz")

	assert.Equal(t, userProvidedServiceInstanceRepo.CreateName, "my-custom-service")
	assert.Equal(t, userProvidedServiceInstanceRepo.CreateParameters, map[string]string{
		"foo": "foo value",
		"bar": "bar value",
		"baz": "baz value",
	})

	assert.Contains(t, fakeUI.Outputs[0], "Creating user provided service")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateUserProvidedServiceWithJson(t *testing.T) {
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	fakeUI := callCreateUserProvidedService(
		[]string{"my-custom-service", `{"foo": "foo value", "bar": "bar value", "baz": "baz value"}`},
		[]string{},
		userProvidedServiceInstanceRepo,
	)

	assert.Empty(t, fakeUI.Prompts)

	assert.Equal(t, userProvidedServiceInstanceRepo.CreateName, "my-custom-service")
	assert.Equal(t, userProvidedServiceInstanceRepo.CreateParameters, map[string]string{
		"foo": "foo value",
		"bar": "bar value",
		"baz": "baz value",
	})

	assert.Contains(t, fakeUI.Outputs[0], "Creating user provided service")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateUserProvidedServiceWithNoSecondArgument(t *testing.T) {
	userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}
	fakeUI := callCreateUserProvidedService(
		[]string{"my-custom-service"},
		[]string{},
		userProvidedServiceInstanceRepo,
	)

	assert.Contains(t, fakeUI.Outputs[0], "FAILED")
}

func callCreateUserProvidedService(args []string, inputs []string, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{Inputs: inputs}
	ctxt := testcmd.NewContext("create-user-provided-service", args)
	cmd := NewCreateUserProvidedService(fakeUI, userProvidedServiceInstanceRepo)
	reqFactory := &testreq.FakeReqFactory{}

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
