package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . SSHCodeActor

type SSHCodeActor interface {
	GetSSHPasscode() (string, error)
}

type SSHCodeCommand struct {
	usage           interface{} `usage:"CF_NAME ssh-code"`
	relatedCommands interface{} `related_commands:"curl, ssh"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SSHCodeActor
}

func (cmd *SSHCodeCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd SSHCodeCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	code, err := cmd.Actor.GetSSHPasscode()
	cmd.UI.DisplayText("{{.SSHCode}}", map[string]interface{}{"SSHCode": code})
	return err
}
