package service_test

import (
	"cf"
	"cf/api"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestCreateService(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{
		cf.ServiceOffering{Label: "cleardb", Plans: []cf.ServicePlan{
			cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"},
		}},
		cf.ServiceOffering{Label: "postgres"},
	}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}
	fakeUI := callCreateService(t,
		[]string{"cleardb", "spark", "my-cleardb-service"},
		[]string{},
		serviceRepo,
	)

	assert.Contains(t, fakeUI.Outputs[0], "Creating service")
	assert.Contains(t, fakeUI.Outputs[0], "my-cleardb-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Equal(t, serviceRepo.CreateServiceInstanceName, "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstancePlan, cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"})
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestCreateServiceWhenServiceAlreadyExists(t *testing.T) {
	serviceOfferings := []cf.ServiceOffering{
		cf.ServiceOffering{Label: "cleardb", Plans: []cf.ServicePlan{
			cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"},
		}},
		cf.ServiceOffering{Label: "postgres"},
	}
	serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings, CreateServiceAlreadyExists: true}
	fakeUI := callCreateService(t,
		[]string{"cleardb", "spark", "my-cleardb-service"},
		[]string{},
		serviceRepo,
	)

	assert.Contains(t, fakeUI.Outputs[0], "Creating service")
	assert.Contains(t, fakeUI.Outputs[0], "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstanceName, "my-cleardb-service")
	assert.Equal(t, serviceRepo.CreateServiceInstancePlan, cf.ServicePlan{Name: "spark", Guid: "cleardb-spark-guid"})
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-cleardb-service")
	assert.Contains(t, fakeUI.Outputs[2], "already exists")
}

func callCreateService(t *testing.T, args []string, inputs []string, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{Inputs: inputs}
	ctxt := testcmd.NewContext("create-service", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewCreateService(fakeUI, config, serviceRepo)
	reqFactory := &testreq.FakeReqFactory{}

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
