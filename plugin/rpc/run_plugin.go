package rpc

import (
	"os"
	"os/exec"

	"code.cloudfoundry.org/cli/v9/cf/configuration/pluginconfig"
)

func RunMethodIfExists(rpcService *CliRpcService, args []string, pluginList map[string]pluginconfig.PluginMetadata) bool {
	for _, metadata := range pluginList {
		for _, command := range metadata.Commands {
			if command.Name == args[0] || command.Alias == args[0] {
				args[0] = command.Name

				rpcService.Start()
				defer rpcService.Stop()

				pluginArgs := append([]string{rpcService.Port()}, args...)

				cmd := exec.Command(metadata.Location, pluginArgs...)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr

				defer stopPlugin(cmd)
				err := cmd.Run()
				if err != nil {
					os.Exit(1)
				}
				return true
			}
		}
	}
	return false
}

func stopPlugin(plugin *exec.Cmd) {
	plugin.Process.Kill()
	plugin.Wait()
}
