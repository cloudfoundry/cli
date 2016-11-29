package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type StartCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	usage               interface{}  `usage:"CF_NAME start APP_NAME"`
	envCFStagingTimeout interface{}  `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}  `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
}

func (_ StartCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ StartCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
