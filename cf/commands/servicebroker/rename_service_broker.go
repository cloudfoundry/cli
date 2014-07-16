package servicebroker

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
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

func (cmd RenameServiceBroker) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-service-broker",
		Description: T("Rename a service broker"),
		Usage:       T("CF_NAME rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER"),
	}
}

func (cmd RenameServiceBroker) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
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

	cmd.ui.Say(T("Renaming service broker {{.OldName}} to {{.NewName}} as {{.Username}}",
		map[string]interface{}{
			"OldName":  terminal.EntityNameColor(serviceBroker.Name),
			"NewName":  terminal.EntityNameColor(c.Args()[1]),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	newName := c.Args()[1]

	apiErr = cmd.repo.Rename(serviceBroker.Guid, newName)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
