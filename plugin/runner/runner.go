package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"code.cloudfoundry.org/cli/v8/cf/commandregistry"
	"code.cloudfoundry.org/cli/v8/cf/configuration/confighelpers"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/trace"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/plugin/rpc"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"

	netrpc "net/rpc"
)

var (
	// ErrFailed is returned when a plugin command fails
	ErrFailed = errors.New("command failed")
	// ParseErr is returned when there's an error parsing plugin arguments
	ParseErr = errors.New("incorrect type for arg")
)

// DisplayUsage interface for commands that can display usage information
type DisplayUsage interface {
	DisplayUsage()
}

// PluginRunner defines the interface for running plugins
type PluginRunner interface {
	Run(args []string) error
}

// pluginRunner implements PluginRunner
type pluginRunner struct {
	config    *configv3.Config
	commandUI *ui.UI
	plugin    configv3.Plugin
}

// NewPluginRunner creates a new PluginRunner instance
func NewPluginRunner(config *configv3.Config, commandUI *ui.UI, plugin configv3.Plugin) PluginRunner {
	return &pluginRunner{
		config:    config,
		commandUI: commandUI,
		plugin:    plugin,
	}
}

// Run executes the plugin with the given arguments
// Based on the plugin execution logic from cf/cmd/cmd.go (lines 140-163)
func (r *pluginRunner) Run(args []string) error {
	// Setup writer and trace logger (mimicking cf/cmd/cmd.go setup)
	var writer io.Writer = os.Stdout
	traceEnv := os.Getenv("CF_TRACE")

	// Get config for trace settings (using legacy config for RPC compatibility)
	configPath, err := confighelpers.DefaultFilePath()
	if err != nil {
		return fmt.Errorf("error getting config path: %w", err)
	}

	errFunc := func(err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config error: %s\n", err)
		}
	}
	legacyConfig := coreconfig.NewRepositoryFromFilepath(configPath, errFunc)
	defer legacyConfig.Close()

	traceConfigVal := legacyConfig.Trace()
	traceLogger := trace.NewLogger(writer, false, traceEnv, traceConfigVal)

	// Create dependencies needed for RPC service
	deps := commandregistry.NewDependency(writer, traceLogger, os.Getenv("CF_DIAL_TIMEOUT"))
	defer deps.Config.Close()

	// Initialize RPC server
	server := netrpc.NewServer()
	rpcService, err := rpc.NewRpcService(
		deps.TeePrinter,
		deps.TeePrinter,
		r.config,
		deps.RepoLocator,
		rpc.NewCommandRunner(),
		deps.Logger,
		writer,
		server,
	)
	if err != nil {
		return fmt.Errorf("error initializing RPC service: %w", err)
	}

	// Start RPC service
	err = rpcService.Start()
	if err != nil {
		return fmt.Errorf("error starting RPC service: %w", err)
	}
	defer rpcService.Stop()

	// Find matching command in plugin and normalize to command name (not alias)
	if len(args) > 0 {
		commandName := args[0]
		for _, pluginCommand := range r.plugin.Commands {
			if pluginCommand.Name == commandName || pluginCommand.Alias == commandName {
				args[0] = pluginCommand.Name
				break
			}
		}
	}

	// Prepare plugin arguments: [port, command, ...args]
	pluginArgs := append([]string{rpcService.Port()}, args...)

	// Execute plugin binary
	cmd := exec.Command(r.plugin.Location, pluginArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// Ensure plugin process is stopped
	defer stopPlugin(cmd)

	// Run the plugin
	err = cmd.Run()
	if err != nil {
		return r.handleError(err)
	}

	return nil
}

// stopPlugin ensures the plugin process is terminated
func stopPlugin(plugin *exec.Cmd) {
	if plugin.Process != nil {
		plugin.Process.Kill()
		plugin.Wait()
	}
}

// handleError processes plugin execution errors and converts them to appropriate error types
// Based on util/plugin/plugin.go handleError
func (r *pluginRunner) handleError(passedErr error) error {
	if passedErr == nil {
		return nil
	}

	translatedErr := translatableerror.ConvertToTranslatableError(passedErr)
	if r.commandUI != nil {
		r.commandUI.DisplayError(translatedErr)
	}

	if _, ok := translatedErr.(DisplayUsage); ok {
		return ParseErr
	}

	return ErrFailed
}
