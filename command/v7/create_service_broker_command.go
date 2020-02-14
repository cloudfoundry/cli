package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/configv3"
)

type CreateServiceBrokerCommand struct {
	BaseCommand

	RequiredArgs    flag.ServiceBrokerArgs `positional-args:"yes"`
	SpaceScoped     bool                   `long:"space-scoped" description:"Make the broker's service plans only visible within the targeted space"`
	usage           interface{}            `usage:"CF_NAME create-service-broker SERVICE_BROKER USERNAME PASSWORD URL [--space-scoped]"`
	relatedCommands interface{}            `related_commands:"enable-service-access, service-brokers, target"`
}

func (cmd *CreateServiceBrokerCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.SpaceScoped, cmd.SpaceScoped)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	var space configv3.Space
	if cmd.SpaceScoped {
		space = cmd.Config.TargetedSpace()
		cmd.UI.DisplayTextWithFlavor(
			"Creating service broker {{.ServiceBroker}} in org {{.Org}} / space {{.Space}} as {{.Username}}...",
			map[string]interface{}{
				"Username":      user.Name,
				"ServiceBroker": cmd.RequiredArgs.ServiceBroker,
				"Org":           cmd.Config.TargetedOrganizationName(),
				"Space":         space.Name,
			},
		)
	} else {
		cmd.UI.DisplayTextWithFlavor(
			"Creating service broker {{.ServiceBroker}} as {{.Username}}...",
			map[string]interface{}{
				"Username":      user.Name,
				"ServiceBroker": cmd.RequiredArgs.ServiceBroker,
			},
		)
	}

	warnings, err := cmd.Actor.CreateServiceBroker(
		v7action.ServiceBrokerModel{
			Name:      cmd.RequiredArgs.ServiceBroker,
			Username:  cmd.RequiredArgs.Username,
			Password:  cmd.RequiredArgs.Password,
			URL:       cmd.RequiredArgs.URL,
			SpaceGUID: space.GUID,
		},
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
