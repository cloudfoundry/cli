package servicebroker

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteServiceBroker struct {
	ui     terminal.UI
	config configuration.Reader
	repo   api.ServiceBrokerRepository
}

func NewDeleteServiceBroker(ui terminal.UI, config configuration.Reader, repo api.ServiceBrokerRepository) (cmd DeleteServiceBroker) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd DeleteServiceBroker) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-service-broker",
		Description: T("Delete a service broker"),
		Usage:       T("CF_NAME delete-service-broker SERVICE_BROKER [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd DeleteServiceBroker) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd DeleteServiceBroker) Run(c *cli.Context) {
	brokerName := c.Args()[0]
	if !c.Bool("f") && !cmd.ui.ConfirmDelete(T("service-broker"), brokerName) {
		return
	}

	cmd.ui.Say(T("Deleting service broker {{.Name}} as {{.Username}}...",
		map[string]interface{}{
			"Name":     terminal.EntityNameColor(brokerName),
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	broker, apiErr := cmd.repo.FindByName(brokerName)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Service Broker {{.Name}} does not exist.", map[string]interface{}{"Name": brokerName}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	apiErr = cmd.repo.Delete(broker.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	return
}
