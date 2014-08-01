package serviceaccess

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DisableServiceAccess struct {
	ui     terminal.UI
	config configuration.Reader
	actor  actors.ServicePlanActor
}

func NewDisableServiceAccess(ui terminal.UI, config configuration.Reader, actor actors.ServicePlanActor) (cmd *DisableServiceAccess) {
	return &DisableServiceAccess{
		ui:     ui,
		config: config,
		actor:  actor,
	}
}

func (cmd *DisableServiceAccess) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *DisableServiceAccess) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "disable-service-access",
		Description: "Set a service to private",
		Usage:       "CF_NAME disable-service-access SERVICE [-p PLAN]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", "name of a particular plan to enable"),
		},
	}
}

func (cmd *DisableServiceAccess) Run(c *cli.Context) {
	serviceName := c.Args()[0]
	planName := c.String("p")

	if planName != "" {

		planOriginallyPublic, err := cmd.actor.UpdateSinglePlanForService(serviceName, planName, false)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}

		if !planOriginallyPublic {
			cmd.ui.Say("Plan %s for service %s is already private", planName, serviceName)
		} else {
			cmd.ui.Say("Disabling access of plan %s for service %s as %s...", planName, serviceName, cmd.config.Username())
		}
	}

	cmd.ui.Say("OK")
}
