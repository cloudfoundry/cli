package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type SSHCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	AppInstanceIndex    int          `long:"app-instance-index" short:"i" description:"Application instance index (Default: 0)"`
	Command             string       `long:"command" short:"c" description:"Command to run. This flag can be defined more than once."`
	DisablePseudoTTY    bool         `long:"disable-pseudo-tty" short:"T" description:"Disable pseudo-tty allocation"`
	ForcePseudoTTY      bool         `long:"force-pseudo-tty" description:"Force pseudo-tty allocation"`
	LocalPort           string       `short:"L" description:"Local port forward specification. This flag can be defined more than once."`
	RemotePseudoTTY     bool         `long:"request-pseudo-tty" short:"t" description:"Request pseudo-tty allocation"`
	SkipHostValidation  bool         `long:"skip-host-validation" short:"k" description:"Skip host key validation"`
	SkipRemoteExecution bool         `long:"skip-remote-execution" short:"N" description:"Do not execute a remote command"`
	usage               interface{}  `usage:"CF_NAME ssh APP_NAME [-i INDEX] [-c COMMAND]... [-L [BIND_ADDRESS:]PORT:HOST:HOST_PORT] [--skip-host-validation] [--skip-remote-execution] [--disable-pseudo-tty | --force-pseudo-tty | --request-pseudo-tty]"`
	relatedCommands     interface{}  `related_commands:"allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled"`
}

func (SSHCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (SSHCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
