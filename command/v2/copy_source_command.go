package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CopySourceCommand struct {
	RequiredArgs        flag.CopySourceArgs `positional-args:"yes"`
	NoRestart           bool                `long:"no-restart" description:"Override restart of the application in target environment after copy-source completes"`
	Organization        string              `short:"o" description:"Org that contains the target application"`
	Space               string              `short:"s" description:"Space that contains the target application"`
	usage               interface{}         `usage:"CF_NAME copy-source SOURCE_APP TARGET_APP [-s TARGET_SPACE [-o TARGET_ORG]] [--no-restart]"`
	relatedCommands     interface{}         `related_commands:"apps, push, restart, target"`
	envCFStagingTimeout interface{}         `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}         `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
}

func (CopySourceCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CopySourceCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
