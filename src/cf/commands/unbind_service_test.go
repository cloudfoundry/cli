package commands_test

import (
	"testhelpers"
	"cf/api"
	"cf/configuration"
	. "cf/commands"
	"testing"
	"cf"
	"github.com/stretchr/testify/assert"
)

func TestUnbindCommand(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}

	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	serviceRepo := &testhelpers.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	config := &configuration.Configuration{}
	fakeUI := callUnbindService([]string{"--service", "my-service", "--app", "my-app"}, config, serviceRepo, appRepo)

	assert.Equal(t, serviceRepo.FindInstanceByNameName, "my-service")
	assert.Equal(t, appRepo.AppName, "my-app")

	assert.Contains(t, fakeUI.Outputs[0], "Unbinding service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-app")

	assert.Equal(t, serviceRepo.UnbindServiceServiceInstance, serviceInstance)
	assert.Equal(t, serviceRepo.UnbindServiceApplication, app)

	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callUnbindService(args []string, config *configuration.Configuration, serviceRepo api.ServiceRepository, appRepo api.ApplicationRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	target := NewUnbindService(fakeUI, config, serviceRepo, appRepo)
	target.Run(testhelpers.NewContext("unbind-service", args))
	return
}
