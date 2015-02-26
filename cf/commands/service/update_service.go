package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/actors/plan_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
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
	if len(c.Args()) != 1 {
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

	serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceInstanceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	planName := c.String("p")

	if planName != "" {
		cmd.ui.Say(T("Updating service instance {{.ServiceName}} as {{.UserName}}...",
			map[string]interface{}{
				"ServiceName": terminal.EntityNameColor(serviceInstanceName),
				"UserName":    terminal.EntityNameColor(cmd.config.Username()),
			}))

		if cmd.config.IsMinApiVersion("2.16.0") {
			err := cmd.updateServiceWithPlan(serviceInstance, planName)
			switch err.(type) {
			case nil:
				err = printSuccessMessageForServiceInstance(serviceInstanceName, cmd.serviceRepo, cmd.ui)
				if err != nil {
					cmd.ui.Failed(err.Error())
				}
			default:
				cmd.ui.Failed(err.Error())
			}
		} else {
			cmd.ui.Failed(T("Updating a plan requires API v{{.RequiredCCAPIVersion}} or newer. Your current target is v{{.CurrentCCAPIVersion}}.",
				map[string]interface{}{
					"RequiredCCAPIVersion": "2.16.0",
					"CurrentCCAPIVersion":  cmd.config.ApiVersion(),
				}))
		}
	} else {
		cmd.ui.Ok()
		cmd.ui.Say(T("No changes were made"))
	}
}

func (cmd *UpdateService) updateServiceWithPlan(serviceInstance models.ServiceInstance, planName string) (err error) {
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

func printSuccessMessageForServiceInstance(serviceInstanceName string, serviceRepo api.ServiceRepository, ui terminal.UI) error {
	instance, apiErr := serviceRepo.FindInstanceByName(serviceInstanceName)
	if apiErr != nil {
		return apiErr
	}

	if instance.ServiceInstanceFields.LastOperation.State == "in progress" {
		ui.Ok()
		ui.Say("")
		ui.Say(T("{{.State}} in progress. Use '{{.ServicesCommand}}' or '{{.ServiceCommand}}' to check operation status.",
			map[string]interface{}{
				"State":           strings.Title(instance.ServiceInstanceFields.LastOperation.Type),
				"ServicesCommand": terminal.CommandColor("cf services"),
				"ServiceCommand":  terminal.CommandColor(fmt.Sprintf("cf service %s", serviceInstanceName)),
			}))
	} else {
		ui.Ok()
	}

	return nil
}
