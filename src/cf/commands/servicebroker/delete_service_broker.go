package servicebroker

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteServiceBroker struct {
	ui     terminal.UI
	config *configuration.Configuration
	repo   api.ServiceBrokerRepository
}

func NewDeleteServiceBroker(ui terminal.UI, config *configuration.Configuration, repo api.ServiceBrokerRepository) (cmd DeleteServiceBroker) {
	cmd.ui = ui
	cmd.config = config
	cmd.repo = repo
	return
}

func (cmd DeleteServiceBroker) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-service-broker")
		return
	}

	reqs = append(reqs, reqFactory.NewLoginRequirement())

	return
}
func (cmd DeleteServiceBroker) Run(c *cli.Context) {
	brokerName := c.Args()[0]
	force := c.Bool("f")

	if !force {
		response := cmd.ui.Confirm(
			"Really delete %s?%s",
			terminal.EntityNameColor(brokerName),
			terminal.PromptColor(">"),
		)
		if !response {
			return
		}
	}

	cmd.ui.Say("Deleting service broker %s as %s...",
		terminal.EntityNameColor(brokerName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	broker, apiResponse := cmd.repo.FindByName(brokerName)

	if apiResponse.IsError() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if apiResponse.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("Service Broker %s does not exist.", brokerName)
		return
	}

	apiResponse = cmd.repo.Delete(broker)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	return
}
