package service_test

import (
	"cf"
	"cf/api"
	. "cf/commands/service"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callBindService(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("bind-service", args)

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

	cmd := NewBindService(fakeUI, config, serviceBindingRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestBindCommand", func() {
			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			serviceInstance := cf.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"
			reqFactory := &testreq.FakeReqFactory{
				Application:     app,
				ServiceInstance: serviceInstance,
			}
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{}
			ui := callBindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), reqFactory.ServiceInstanceName, "my-service")

			assert.Equal(mr.T(), len(ui.Outputs), 3)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				{"OK"},
				{"TIP"},
			})
			assert.Equal(mr.T(), serviceBindingRepo.CreateServiceInstanceGuid, "my-service-guid")
			assert.Equal(mr.T(), serviceBindingRepo.CreateApplicationGuid, "my-app-guid")
		})
		It("TestBindCommandIfServiceIsAlreadyBound", func() {

			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			serviceInstance := cf.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"
			reqFactory := &testreq.FakeReqFactory{
				Application:     app,
				ServiceInstance: serviceInstance,
			}
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{CreateErrorCode: "90003"}
			ui := callBindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

			assert.Equal(mr.T(), len(ui.Outputs), 3)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Binding service"},
				{"OK"},
				{"my-app", "is already bound", "my-service"},
			})
		})
		It("TestBindCommandFailsWithUsage", func() {

			reqFactory := &testreq.FakeReqFactory{}
			serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

			ui := callBindService(mr.T(), []string{"my-service"}, reqFactory, serviceBindingRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callBindService(mr.T(), []string{"my-app"}, reqFactory, serviceBindingRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callBindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
	})
}
