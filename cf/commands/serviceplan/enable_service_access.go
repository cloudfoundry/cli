package serviceplan

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type EnableServiceAccess struct {
	ui     terminal.UI
	config configuration.Reader
	actor  actors.ServiceActor
}

func NewEnableServiceAccess(ui terminal.UI, config configuration.Reader, actor actors.ServiceActor) *EnableServiceAccess {
	return &EnableServiceAccess{
		ui:     ui,
		config: config,
		actor:  actor,
	}
}

func (cmd *EnableServiceAccess) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *EnableServiceAccess) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "enable-service-access",
		Description: "Set a service to public",
		Usage:       "CF_NAME enable-service-access",
		Flags:       []cli.Flag{},
	}
}

func (cmd *EnableServiceAccess) Run(c *cli.Context) {
	serviceName := context.Args()[0]

	service, err := cmd.actor.GetService(serviceName)
	if err != nil {
		cmd.ui.Failed("Could not find service plan.\n %s", err)
	}

	cmd.ui.Say("OK")
}
