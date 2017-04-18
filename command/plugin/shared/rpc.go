package shared

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	netrpc "net/rpc"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin/rpc"
)

type Config interface {
	DialTimeout() time.Duration
	Verbose() (bool, []string)
}

type UI interface {
	Writer() io.Writer
}

func NewRPCService(config Config, ui UI) (*rpc.CliRpcService, error) {
	isVerbose, logFiles := config.Verbose()
	traceLogger := trace.NewLogger(ui.Writer(), isVerbose, logFiles...)

	deps := commandregistry.NewDependency(ui.Writer(), traceLogger, fmt.Sprint(config.DialTimeout().Seconds()))
	defer deps.Config.Close()

	server := netrpc.NewServer()
	return rpc.NewRpcService(deps.TeePrinter, deps.TeePrinter, deps.Config, deps.RepoLocator, rpc.NewCommandRunner(), deps.Logger, ui.Writer(), server)
}

type PluginUninstaller struct {
	config Config
	ui     UI
}

func NewPluginUninstaller(config Config, ui UI) *PluginUninstaller {
	return &PluginUninstaller{
		config: config,
		ui:     ui,
	}
}

func (p PluginUninstaller) Uninstall(location string) error {
	rpcService, err := NewRPCService(p.config, p.ui)
	if err != nil {
		return err
	}

	err = rpcService.Start()
	if err != nil {
		return err
	}
	defer rpcService.Stop()

	pluginInvocation := exec.Command(location, rpcService.Port(), "CLI-MESSAGE-UNINSTALL")
	pluginInvocation.Stdout = os.Stdout
	pluginInvocation.Stderr = os.Stderr

	return pluginInvocation.Run()
}
