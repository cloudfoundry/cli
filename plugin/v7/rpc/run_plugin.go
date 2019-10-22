// +build V7

package rpc

import (
	"os"
	"os/exec"

	"code.cloudfoundry.org/cli/util/configv3"
)

func RunMethodIfExists(rpcService *CliRpcService, args []string, plugin configv3.Plugin) {
	err := rpcService.Start()
	if err != nil {
		os.Exit(1)
	}

	defer rpcService.Stop()

	pluginArgs := append([]string{rpcService.Port()}, args...)

	cmd := exec.Command(plugin.Location, pluginArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	defer stopPlugin(cmd)

	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

func stopPlugin(plugin *exec.Cmd) {
	err := plugin.Process.Kill()
	if err != nil {
		os.Exit(1)
	}

	err = plugin.Wait()
	if err != nil {
		os.Exit(1)
	}
}
