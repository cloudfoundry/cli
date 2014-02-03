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

func callCreateService(t mr.TestingT, args []string, inputs []string, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{Inputs: inputs}
	ctxt := testcmd.NewContext("create-service", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewCreateService(fakeUI, config, serviceRepo)
	reqFactory := &testreq.FakeReqFactory{}

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateService", func() {
			offering := cf.ServiceOffering{}
			offering.Label = "cleardb"
			plan := cf.ServicePlanFields{}
			plan.Name = "spark"
			plan.Guid = "cleardb-spark-guid"
			offering.Plans = []cf.ServicePlanFields{plan}
			offering2 := cf.ServiceOffering{}
			offering2.Label = "postgres"
			serviceOfferings := []cf.ServiceOffering{offering, offering2}
			serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings}
			ui := callCreateService(mr.T(), []string{"cleardb", "spark", "my-cleardb-service"},
				[]string{},
				serviceRepo,
			)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating service", "my-cleardb-service", "my-org", "my-space", "my-user"},
				{"OK"},
			})
			assert.Equal(mr.T(), serviceRepo.CreateServiceInstanceName, "my-cleardb-service")
			assert.Equal(mr.T(), serviceRepo.CreateServiceInstancePlanGuid, "cleardb-spark-guid")
		})
		It("TestCreateServiceWhenServiceAlreadyExists", func() {

			offering := cf.ServiceOffering{}
			offering.Label = "cleardb"
			plan := cf.ServicePlanFields{}
			plan.Name = "spark"
			plan.Guid = "cleardb-spark-guid"
			offering.Plans = []cf.ServicePlanFields{plan}
			offering2 := cf.ServiceOffering{}
			offering2.Label = "postgres"
			serviceOfferings := []cf.ServiceOffering{offering, offering2}
			serviceRepo := &testapi.FakeServiceRepo{ServiceOfferings: serviceOfferings, CreateServiceAlreadyExists: true}
			ui := callCreateService(mr.T(), []string{"cleardb", "spark", "my-cleardb-service"},
				[]string{},
				serviceRepo,
			)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating service", "my-cleardb-service"},
				{"OK"},
				{"my-cleardb-service", "already exists"},
			})
			assert.Equal(mr.T(), serviceRepo.CreateServiceInstanceName, "my-cleardb-service")
			assert.Equal(mr.T(), serviceRepo.CreateServiceInstancePlanGuid, "cleardb-spark-guid")
		})
	})
}
