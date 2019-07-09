package v7

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

//go:generate counterfeiter . CancelDeploymentActor

type CancelDeploymentActor interface {
}

type CancelDeploymentCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	usage           interface{}  `usage:"CF_NAME cancel-deployment APP_NAME\n\nEXAMPLES:\n   cf cancel-deployment my-app"`
	relatedCommands interface{}  `related_commands:"app, push"`

	UI          command.UI
	Config      command.Config
	Actor       CancelDeploymentActor
	SharedActor command.SharedActor
}

func (cmd *CancelDeploymentCommand) Execute() {}
