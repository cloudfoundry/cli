package serviceplan

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type EnableServiceAccess struct {
	ui     terminal.UI
	config configuration.Reader
	actor  actors.ServicePlanActor
}

func NewEnableServiceAccess(ui terminal.UI, config configuration.Reader, actor actors.ServicePlanActor) (cmd *EnableServiceAccess) {
	return &EnableServiceAccess{
		ui:     ui,
		config: config,
		actor:  actor,
	}
}

func (cmd *EnableServiceAccess) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *EnableServiceAccess) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "enable-service-access",
		Description: "Set a service to public",
		Usage:       "CF_NAME enable-service-access [-p PLAN]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", "name of a particular plan to enable"),
		},
	}
}

func (cmd *EnableServiceAccess) Run(c *cli.Context) {
	planToFilter := c.String("p")
	serviceName := c.Args()[0]

	var err error
	var plan models.ServicePlanFields

	if planToFilter != "" {
		plan, err = cmd.actor.GetSingleServicePlanForService(serviceName, planToFilter)
		if err != nil {
			cmd.ui.Failed("Could not find service plan.\n%s", err)
		}

	}

	if plan.Public {
		cmd.ui.Say("Plan %s for service %s is already public", planToFilter, serviceName)
	}

	cmd.ui.Say("OK")
}
