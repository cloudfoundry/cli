package service_test

import (
	"cf/api"
	. "cf/commands/service"
	"cf/configuration"
	"cf/models"
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

func callUpdateUserProvidedService(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("update-user-provided-service", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := models.OrganizationFields{}
	org.Name = "my-org"
	space := models.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewUpdateUserProvidedService(fakeUI, config, userProvidedServiceInstanceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUpdateUserProvidedServiceFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

			ui := callUpdateUserProvidedService(mr.T(), []string{}, reqFactory, userProvidedServiceInstanceRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUpdateUserProvidedService(mr.T(), []string{"foo"}, reqFactory, userProvidedServiceInstanceRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestUpdateUserProvidedServiceRequirements", func() {

			args := []string{"service-name"}
			reqFactory := &testreq.FakeReqFactory{}
			userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

			reqFactory.LoginSuccess = false
			callUpdateUserProvidedService(mr.T(), args, reqFactory, userProvidedServiceInstanceRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callUpdateUserProvidedService(mr.T(), args, reqFactory, userProvidedServiceInstanceRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			assert.Equal(mr.T(), reqFactory.ServiceInstanceName, "service-name")
		})
		It("TestUpdateUserProvidedServiceWhenNoFlagsArePresent", func() {

			args := []string{"service-name"}
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "found-service-name"
			reqFactory := &testreq.FakeReqFactory{
				LoginSuccess:    true,
				ServiceInstance: serviceInstance,
			}
			repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
			ui := callUpdateUserProvidedService(mr.T(), args, reqFactory, repo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Updating user provided service", "found-service-name", "my-org", "my-space", "my-user"},
				{"OK"},
				{"No changes"},
			})
		})
		It("TestUpdateUserProvidedServiceWithJson", func() {

			args := []string{"-p", `{"foo":"bar"}`, "-l", "syslog://example.com", "service-name"}
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "found-service-name"
			reqFactory := &testreq.FakeReqFactory{
				LoginSuccess:    true,
				ServiceInstance: serviceInstance,
			}
			repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
			ui := callUpdateUserProvidedService(mr.T(), args, reqFactory, repo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Updating user provided service", "found-service-name", "my-org", "my-space", "my-user"},
				{"OK"},
				{"TIP"},
			})
			assert.Equal(mr.T(), repo.UpdateServiceInstance.Name, serviceInstance.Name)
			assert.Equal(mr.T(), repo.UpdateServiceInstance.Params, map[string]string{"foo": "bar"})
			assert.Equal(mr.T(), repo.UpdateServiceInstance.SysLogDrainUrl, "syslog://example.com")
		})
		It("TestUpdateUserProvidedServiceWithoutJson", func() {

			args := []string{"-l", "syslog://example.com", "service-name"}
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "found-service-name"
			reqFactory := &testreq.FakeReqFactory{
				LoginSuccess:    true,
				ServiceInstance: serviceInstance,
			}
			repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
			ui := callUpdateUserProvidedService(mr.T(), args, reqFactory, repo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Updating user provided service"},
				{"OK"},
			})
		})
		It("TestUpdateUserProvidedServiceWithInvalidJson", func() {

			args := []string{"-p", `{"foo":"ba`, "service-name"}
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "found-service-name"
			reqFactory := &testreq.FakeReqFactory{
				LoginSuccess:    true,
				ServiceInstance: serviceInstance,
			}
			userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

			ui := callUpdateUserProvidedService(mr.T(), args, reqFactory, userProvidedServiceInstanceRepo)

			assert.NotEqual(mr.T(), userProvidedServiceInstanceRepo.UpdateServiceInstance, serviceInstance)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"JSON is invalid"},
			})
		})
		It("TestUpdateUserProvidedServiceWithAServiceInstanceThatIsNotUserProvided", func() {

			args := []string{"-p", `{"foo":"bar"}`, "service-name"}
			plan := models.ServicePlanFields{}
			plan.Guid = "my-plan-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "found-service-name"
			serviceInstance.ServicePlan = plan

			reqFactory := &testreq.FakeReqFactory{
				LoginSuccess:    true,
				ServiceInstance: serviceInstance,
			}
			userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

			ui := callUpdateUserProvidedService(mr.T(), args, reqFactory, userProvidedServiceInstanceRepo)

			assert.NotEqual(mr.T(), userProvidedServiceInstanceRepo.UpdateServiceInstance, serviceInstance)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Service Instance is not user provided"},
			})
		})
	})
}
