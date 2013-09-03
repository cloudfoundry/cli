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

func TestBindCommand(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}

	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	serviceRepo := &testhelpers.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	config := &configuration.Configuration{}
	fakeUI := callBindService([]string{"--service", "my-service", "--app", "my-app"}, config, serviceRepo, appRepo)

	assert.Equal(t, serviceRepo.FindInstanceByNameName, "my-service")
	assert.Equal(t, appRepo.AppName, "my-app")

	assert.Contains(t, fakeUI.Outputs[0], "Binding service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")

	assert.Equal(t, serviceRepo.BindServiceServiceInstance, serviceInstance)
	assert.Equal(t, serviceRepo.BindServiceApplication, app)

	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callBindService(args []string, config *configuration.Configuration, serviceRepo api.ServiceRepository, appRepo api.ApplicationRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	target := NewBindService(fakeUI, config, serviceRepo, appRepo)
	target.Run(testhelpers.NewContext("bind-service", args))
	return
}
