package servicebroker

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ListServiceBrokers struct {
	ui     terminal.UI
	config configuration.Reader
	repo   api.ServiceBrokerRepository
}

func NewListServiceBrokers(ui terminal.UI, config configuration.Reader, repo api.ServiceBrokerRepository) (cmd ListServiceBrokers) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd ListServiceBrokers) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "service-brokers",
		Description: T("List service brokers"),
		Usage:       "CF_NAME service-brokers",
	}
}

func (cmd ListServiceBrokers) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd ListServiceBrokers) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting service brokers as {{.Username}}...\n",
		map[string]interface{}{
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	table := cmd.ui.Table([]string{T("name"), T("url")})
	foundBrokers := false
	apiErr := cmd.repo.ListServiceBrokers(func(serviceBroker models.ServiceBroker) bool {
		table.Add(serviceBroker.Name, serviceBroker.Url)
		foundBrokers = true
		return true
	})
	table.Print()

	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching service brokers.\n{{.Error}}", map[string]interface{}{"Error": apiErr}))
		return
	}

	if !foundBrokers {
		cmd.ui.Say(T("No service brokers found"))
	}
}
