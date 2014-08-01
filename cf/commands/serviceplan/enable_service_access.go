package serviceplan

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
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
		Usage:       "CF_NAME enable-service-access SERVICE [-p PLAN]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", "name of a particular plan to enable"),
		},
	}
}

func (cmd *EnableServiceAccess) Run(c *cli.Context) {
	serviceName := c.Args()[0]
	planName := c.String("p")

	if planName != "" {

		planOriginallyPublic, err := cmd.actor.UpdateSinglePlanForService(serviceName, planName)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}

		if planOriginallyPublic {
			cmd.ui.Say("Plan %s for service %s is already public", planName, serviceName)
		} else {
			cmd.ui.Say("Enabling access of plan %s for service %s", planName, serviceName)
		}
	}

	cmd.ui.Say("OK")
}
