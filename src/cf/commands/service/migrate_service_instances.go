package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
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
		ServiceLabel:    c.Args()[0],
		ServiceProvider: c.Args()[1],
		ServicePlanName: c.Args()[2],
	}
	v2 := api.ServicePlanDescription{
		ServiceLabel:    c.Args()[3],
		ServicePlanName: c.Args()[4],
	}
	force := c.Bool("f")

	v1Guid, apiErr := cmd.serviceRepo.FindServicePlanByDescription(v1)
	if apiErr != nil {
		if apiErr.IsNotFound() {
			cmd.ui.Failed("Plan %s cannot be found", terminal.EntityNameColor(v1.String()))
		} else {
			cmd.ui.Failed(apiErr.Error())
		}
		return
	}

	v2Guid, apiErr := cmd.serviceRepo.FindServicePlanByDescription(v2)
	if apiErr != nil {
		if apiErr.IsNotFound() {
			cmd.ui.Failed("Plan %s cannot be found", terminal.EntityNameColor(v2.String()))
		} else {
			cmd.ui.Failed(apiErr.Error())
		}
		return
	}

	count, apiErr := cmd.serviceRepo.GetServiceInstanceCountForServicePlan(v1Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	} else if count == 0 {
		cmd.ui.Failed("Plan %s has no service instances to migrate", terminal.EntityNameColor(v1.String()))
		return
	}

	cmd.ui.Warn("WARNING: This operation is internal to Cloud Foundry; service brokers will not be contacted and" +
		" resources for service instances will not be altered. The primary use case for this operation is" +
		" to replace a service broker which implements the v1 Service Broker API with a broker which" +
		" implements the v2 API by remapping service instances from v1 plans to v2 plans.  We recommend" +
		" making the v1 plan private or shutting down the v1 broker to prevent additional instances from" +
		" being created. Once service instances have been migrated, the v1 services and plans can be" +
		" removed from Cloud Foundry.")

	serviceInstancesPhrase := pluralizeServiceInstances(count)

	if !force {
		response := cmd.ui.Confirm("Really migrate %s from plan %s to %s?>",
			serviceInstancesPhrase,
			terminal.EntityNameColor(v1.String()),
			terminal.EntityNameColor(v2.String()),
		)
		if !response {
			return
		}
	}

	cmd.ui.Say("Attempting to migrate %s...", serviceInstancesPhrase)

	changedCount, apiErr := cmd.serviceRepo.MigrateServicePlanFromV1ToV2(v1Guid, v2Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Say("%s migrated.", pluralizeServiceInstances(changedCount))

	cmd.ui.Ok()

	return
}

func pluralizeServiceInstances(count int) string {
	var phrase string
	if count == 1 {
		phrase = "service instance"
	} else {
		phrase = "service instances"
	}

	return fmt.Sprintf("%d %s", count, phrase)
}
