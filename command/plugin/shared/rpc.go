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
	"code.cloudfoundry.org/cli/util/configv3"
)

type Config interface {
	DialTimeout() time.Duration
	Verbose() (bool, []string)
}

type UI interface {
	Writer() io.Writer
}

type RPCService struct {
	config     Config
	ui         UI
	rpcService *rpc.CliRpcService
}

func NewRPCService(config Config, ui UI) (*RPCService, error) {
	isVerbose, logFiles := config.Verbose()
	traceLogger := trace.NewLogger(ui.Writer(), isVerbose, logFiles...)

	deps := commandregistry.NewDependency(ui.Writer(), traceLogger, fmt.Sprint(config.DialTimeout().Seconds()))
	defer deps.Config.Close()

	server := netrpc.NewServer()
	rpcService, err := rpc.NewRpcService(deps.TeePrinter, deps.TeePrinter, deps.Config, deps.RepoLocator, rpc.NewCommandRunner(), deps.Logger, ui.Writer(), server)
	if err != nil {
		return nil, err
	}

	return &RPCService{
		config:     config,
		ui:         ui,
		rpcService: rpcService,
	}, nil
}

func (r RPCService) Run(path string, command string) error {
	err := r.rpcService.Start()
	if err != nil {
		return err
	}
	defer r.rpcService.Stop()

	cmd := exec.Command(path, r.rpcService.Port(), command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (r RPCService) GetMetadata(path string) (configv3.Plugin, error) {
	err := r.Run(path, "SendMetadata")
	if err != nil {
		return configv3.Plugin{}, err
	}

	metadata := r.rpcService.RpcCmd.PluginMetadata
	plugin := configv3.Plugin{
		Name: metadata.Name,
		Version: configv3.PluginVersion{
			Major: metadata.Version.Major,
			Minor: metadata.Version.Minor,
			Build: metadata.Version.Build,
		},
		Commands: make([]configv3.PluginCommand, len(metadata.Commands)),
	}

	for i, command := range metadata.Commands {
		plugin.Commands[i] = configv3.PluginCommand{
			Name:     command.Name,
			Alias:    command.Alias,
			HelpText: command.HelpText,
			UsageDetails: configv3.PluginUsageDetails{
				Usage:   command.UsageDetails.Usage,
				Options: command.UsageDetails.Options,
			},
		}
	}

	return plugin, nil
}
