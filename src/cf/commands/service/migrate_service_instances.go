package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type MigrateServiceInstances struct {
	ui          terminal.UI
	configRepo  configuration.Reader
	serviceRepo api.ServiceRepository
}

func NewMigrateServiceInstances(ui terminal.UI, configRepo configuration.Reader, serviceRepo api.ServiceRepository) (cmd *MigrateServiceInstances) {
	cmd = new(MigrateServiceInstances)
	cmd.ui = ui
	cmd.configRepo = configRepo
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd *MigrateServiceInstances) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 5 {
		err = errors.New("incorrect usage")
		cmd.ui.FailWithUsage(c, "migrate-service-instances")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *MigrateServiceInstances) Run(c *cli.Context) {
	v1 := api.V1ServicePlanDescription{
		ServiceName:     c.Args()[0],
		ServiceProvider: c.Args()[1],
		ServicePlanName: c.Args()[2],
	}
	v2 := api.V2ServicePlanDescription{
		ServiceName:     c.Args()[3],
		ServicePlanName: c.Args()[4],
	}
	cmd.ui.Say("Migrating plan '%s' for service '%s' from provider '%s' to plan '%s' with service '%s'",
		v1.ServicePlanName,
		v1.ServiceName,
		v1.ServiceProvider,
		v2.ServicePlanName,
		v2.ServiceName,
	)
	v1Guid, v2Guid, apiResponse := cmd.serviceRepo.FindServicePlanToMigrateByDescription(v1, v2)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	apiResponse = cmd.serviceRepo.MigrateServicePlanFromV1ToV2(v1Guid, v2Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
	}
	cmd.ui.Say("OK")

	return
}
