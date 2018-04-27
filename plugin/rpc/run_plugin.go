package rpc

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
)

func RunMethodIfExists(rpcService *CliRpcService, args []string, pluginList map[string]pluginconfig.PluginMetadata) bool {
	for _, metadata := range pluginList {
		for _, command := range metadata.Commands {
			if command.Name == args[0] || command.Alias == args[0] {
				args[0] = command.Name

				rpcService.Start()
				defer rpcService.Stop()

				pluginArgs := append([]string{rpcService.Port()}, args...)

				go func() {
					sig := make(chan os.Signal, 3)
					signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
					for {
						<-sig
					}
				}()

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
