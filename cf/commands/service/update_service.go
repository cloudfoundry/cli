package service

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/actors/plan_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type UpdateService struct {
	ui          terminal.UI
	config      core_config.Reader
	serviceRepo api.ServiceRepository
	planBuilder plan_builder.PlanBuilder
}

func NewUpdateService(ui terminal.UI, config core_config.Reader, serviceRepo api.ServiceRepository, planBuilder plan_builder.PlanBuilder) (cmd *UpdateService) {
	return &UpdateService{
		ui:          ui,
		config:      config,
		serviceRepo: serviceRepo,
		planBuilder: planBuilder,
	}
}

func (cmd *UpdateService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "update-service",
		Description: T("Update a service instance"),
		Usage:       T("CF_NAME update-service SERVICE [-p NEW_PLAN]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", T("Change service plan for a service instance")),
		},
	}
}

func (cmd *UpdateService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return
}

func (cmd *UpdateService) Run(c *cli.Context) {
	serviceInstanceName := c.Args()[0]
	planName := c.String("p")

	if planName != "" {
		cmd.ui.Say(T("Updating service instance {{.ServiceName}} as {{.UserName}}...",
			map[string]interface{}{
				"ServiceName": terminal.EntityNameColor(serviceInstanceName),
				"UserName":    terminal.EntityNameColor(cmd.config.Username()),
			}))

		err := cmd.updateServiceWithPlan(serviceInstanceName, planName)
		switch err.(type) {
		case nil:
			cmd.ui.Ok()
		default:
			cmd.ui.Failed(err.Error())
		}
	} else {
		cmd.ui.Ok()
		cmd.ui.Say(T("No changes were made"))
	}
}

func (cmd *UpdateService) updateServiceWithPlan(serviceInstanceName, planName string) (err error) {
	serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceInstanceName)
	if err != nil {
		return
	}

	plans, err := cmd.planBuilder.GetPlansForServiceForOrg(serviceInstance.ServiceOffering.Guid, cmd.config.OrganizationFields().Name)
	if err != nil {
		return
	}

	for _, plan := range plans {
		if plan.Name == planName {
			err = cmd.serviceRepo.UpdateServiceInstance(serviceInstance.Guid, plan.Guid)
			return
		}
	}
	err = errors.New(T("Plan does not exist for the {{.ServiceName}} service",
		map[string]interface{}{"ServiceName": serviceInstance.ServiceOffering.Label}))

	return
}
