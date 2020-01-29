// +build V7

package plugin_transition

import (
	netrpc "net/rpc"
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/plugin/v7/rpc"
	"code.cloudfoundry.org/cli/util/command_parser"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/clock"
)

func RunPlugin(plugin configv3.Plugin, ui command.UI) error {
	config, err := configv3.LoadConfig(configv3.FlagOverride{
		Verbose: common.Commands.VerboseOrVersion,
	})

	if err != nil {
		if _, ok := err.(translatableerror.EmptyConfigError); !ok {
			return err
		}
	}

	sharedActor := sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}

	v7Actor := v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	cmdParser, err := command_parser.NewCommandParser()
	if err != nil {
		return err
	}

	server := netrpc.NewServer()
	rpcService, err := rpc.NewRpcService(nil, server, config, v7Actor, &cmdParser)

	if err != nil {
		return err
	}

	rpc.RunMethod(rpcService, os.Args[1:], plugin)
	return nil
}
