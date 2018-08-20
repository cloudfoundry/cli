package v2

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . DeleteBuildpackActor

type DeleteBuildpackActor interface {
	CloudControllerAPIVersion() string
}

type DeleteBuildpackCommand struct {
	RequiredArgs    flag.BuildpackName `positional-args:"yes"`
	Force           bool               `short:"f" description:"Force deletion without confirmation"`
	Stack           string             `short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	usage           interface{}        `usage:"CF_NAME delete-buildpack BUILDPACK [-f] [-s STACK]"`
	relatedCommands interface{}        `related_commands:"buildpacks"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DeleteBuildpackActor
}

func (cmd *DeleteBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	return nil
}

func (cmd DeleteBuildpackCommand) Execute(args []string) error {
	if cmd.stackSpecified() {
		err := command.MinimumCCAPIVersionCheck(
			cmd.Actor.CloudControllerAPIVersion(),
			ccversion.MinVersionBuildpackStackAssociationV2,
			"Option `-s`",
		)
		if err != nil {
			return err
		}
	}

	return translatableerror.UnrefactoredCommandError{}
}

func (cmd DeleteBuildpackCommand) stackSpecified() bool {
	return len(cmd.Stack) > 0
}
