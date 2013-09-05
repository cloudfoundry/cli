package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateService(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{
		cf.ServiceOffering{Label: "cleardb", Plans: []cf.ServicePlan{
			cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"},
		}},
		cf.ServiceOffering{Label: "postgres"},
	}
	serviceRepo := &testhelpers.FakeServiceRepo{ServiceOfferings: serviceOfferings}
	config := &configuration.Configuration{}
	fakeUI := callCreateService(
		[]string{"--offering", "cleardb", "--plan", "spark", "--name", "my-cleardb-service"},
		[]string{},
		config,
		serviceRepo,
	)

	assert.Contains(t, fakeUI.Outputs[0], "Creating service")
	assert.Contains(t, fakeUI.Outputs[0], "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstanceName, "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstancePlan, cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"})
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateUserProvidedService(t *testing.T) {
	serviceRepo := &testhelpers.FakeServiceRepo{}
	config := &configuration.Configuration{}
	fakeUI := callCreateService(
		[]string{"--offering", "user-provided", "--name", "my-custom-service", "--parameters", `"foo, bar, baz"`},
		[]string{"foo value", "bar value", "baz value"},
		config,
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

	assert.Contains(t, fakeUI.Outputs[0], "Creating service")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateUserProvidedServiceWithNoParameterList(t *testing.T) {
	serviceRepo := &testhelpers.FakeServiceRepo{}
	config := &configuration.Configuration{}
	fakeUI := callCreateService(
		[]string{"--offering", "user-provided", "--name", "my-custom-service"},
		[]string{},
		config,
		serviceRepo,
	)

	assert.Contains(t, fakeUI.Outputs[0], "FAILED")
}

func callCreateService(args []string, inputs []string, config *configuration.Configuration, serviceRepo api.ServiceRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = &testhelpers.FakeUI{Inputs: inputs}
	ctxt := testhelpers.NewContext("create-service", args)
	cmd := NewCreateService(fakeUI, config, serviceRepo)
	reqFactory := &testhelpers.FakeReqFactory{}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
