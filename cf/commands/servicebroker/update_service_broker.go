package servicebroker

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type UpdateServiceBroker struct {
	ui     terminal.UI
	config configuration.Reader
	repo   api.ServiceBrokerRepository
}

func NewUpdateServiceBroker(ui terminal.UI, config configuration.Reader, repo api.ServiceBrokerRepository) (cmd UpdateServiceBroker) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd UpdateServiceBroker) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "update-service-broker",
		Description: T("Update a service broker"),
		Usage:       T("CF_NAME update-service-broker SERVICE_BROKER [-u USERNAME] [-p PASSWORD] [--url URL]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("u", T("Username")),
			flag_helpers.NewStringFlag("p", T("Password")),
			flag_helpers.NewStringFlag("url", T("URL")),
		},
	}
}

func (cmd UpdateServiceBroker) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())

	return
}

func (cmd UpdateServiceBroker) Run(c *cli.Context) {
	serviceBroker, apiErr := cmd.repo.FindByName(c.Args()[0])
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Say(T("Updating service broker {{.Name}} as {{.Username}}...",
		map[string]interface{}{
			"Name":     terminal.EntityNameColor(serviceBroker.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	// We only want to update the passed in flags. The repo does not send a zero value.
	serviceBroker.Username = c.String("u")
	serviceBroker.Password = c.String("p")
	serviceBroker.Url = c.String("url")

	apiErr = cmd.repo.Update(serviceBroker)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
