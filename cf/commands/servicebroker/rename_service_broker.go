package servicebroker

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type RenameServiceBroker struct {
	ui     terminal.UI
	config configuration.Reader
	repo   api.ServiceBrokerRepository
}

func NewRenameServiceBroker(ui terminal.UI, config configuration.Reader, repo api.ServiceBrokerRepository) (cmd RenameServiceBroker) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (command RenameServiceBroker) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-service-broker",
		Description: "Rename a service broker",
		Usage:       "CF_NAME rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER",
	}
}

func (cmd RenameServiceBroker) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename-service-broker")
		return
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())

	return
}

func (cmd RenameServiceBroker) Run(c *cli.Context) {
	serviceBroker, apiErr := cmd.repo.FindByName(c.Args()[0])
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Say("Renaming service broker %s to %s as %s",
		terminal.EntityNameColor(serviceBroker.Name),
		terminal.EntityNameColor(c.Args()[1]),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	newName := c.Args()[1]

	apiErr = cmd.repo.Rename(serviceBroker.Guid, newName)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
