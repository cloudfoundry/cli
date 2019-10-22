// +build V7

package plugin_transition

import (
	netrpc "net/rpc"
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/plugin/v7/rpc"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/clock"
)

func RunPlugin(plugin configv3.Plugin) {
	config, err := configv3.LoadConfig(configv3.FlagOverride{
		Verbose: common.Commands.VerboseOrVersion,
	})

	if err != nil {
		if _, ok := err.(translatableerror.EmptyConfigError); !ok {
			panic(err)
		}
	}

	sharedActor := sharedaction.NewActor(config)

	//UI for the actor is used for logging. This is probably unnecessay for plugin's use of an actor
	// ui, err := ui.NewUI(config)
	// if err != nil {
	// 	panic(err)
	// }
	// defer ui.FlushDeferred()

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, nil, "")
	if err != nil {
		panic(err)
	}

	v7Actor := v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	server := netrpc.NewServer()
	rpcService, err := rpc.NewRpcService(nil, nil, nil, server, config, v7Actor)

	if err != nil {
		panic(err)
	}

	rpc.RunMethodIfExists(rpcService, os.Args[1:], plugin)
}
