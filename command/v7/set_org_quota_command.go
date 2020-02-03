package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type SetOrgQuotaActor interface {
}

type SetOrgQuotaCommand struct {
	RequiredArgs    flag.SetOrgQuotaArgs `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME set-quota ORG QUOTA\n\nTIP:\n   View allowable quotas with 'CF_NAME quotas'"`
	relatedCommands interface{}          `related_commands:"orgs, quotas"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor           SetOrgQuotaActor
}

func (cmd SetOrgQuotaCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)

	return nil
}

func (SetOrgQuotaCommand) Execute(args []string) error {
	//get quota
	//return if err
	//get org
	//return if err
	//create relationship
	//return if err
	// ????????
	return nil
}
