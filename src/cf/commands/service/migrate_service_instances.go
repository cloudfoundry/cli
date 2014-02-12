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
	v1 := api.ServicePlanDescription{
		ServiceName:     c.Args()[0],
		ServiceProvider: c.Args()[1],
		ServicePlanName: c.Args()[2],
	}
	v2 := api.ServicePlanDescription{
		ServiceName:     c.Args()[3],
		ServicePlanName: c.Args()[4],
	}

	v1Guid, apiResponse := cmd.serviceRepo.FindServicePlanByDescription(v1)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	v2Guid, apiResponse := cmd.serviceRepo.FindServicePlanByDescription(v2)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	count, apiResponse := cmd.serviceRepo.GetServiceInstanceCountForServicePlan(v1Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	} else if count == 0 {
		cmd.ui.Failed("Plan %s has no service instances to migrate", v1)
		return
	}

	cmd.ui.Warn("WARNING: This operation is internal to Cloud Foundry; service brokers will not be contacted and" +
		" resources for service instances will not be altered. The primary use case for this operation is" +
		" to replace a service broker which implements the v1 Service Broker API with a broker which" +
		" implements the v2 API by remapping service instances from v1 plans to v2 plans.  We recommend" +
		" making the v1 plan private or shutting down the v1 broker to prevent additional instances from" +
		" being created. Once service instances have been migrated, the v1 services and plans can be" +
		" removed from Cloud Foundry.\n\n")
	var serviceInstancesPhrase string
	if count == 1 {
		serviceInstancesPhrase = "service instance"
	} else {
		serviceInstancesPhrase = "service instances"
	}
	response := cmd.ui.Confirm("Really migrate %d %s from plan %s to %s?", count, serviceInstancesPhrase, v1, v2)
	if !response {
		return
	}

	cmd.ui.Say("Migrating %d %s...", count, serviceInstancesPhrase)

	apiResponse = cmd.serviceRepo.MigrateServicePlanFromV1ToV2(v1Guid, v2Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
	}
	cmd.ui.Ok()

	return
}
