package runner

import (
	"errors"
	"fmt"
	netrpc "net/rpc"
	"os"
	"os/exec"
	"strings"

	"code.cloudfoundry.org/cli/v8/actor/sharedaction"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/command/v7/shared"
	"code.cloudfoundry.org/cli/v8/plugin/rpc"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	"code.cloudfoundry.org/clock"
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
	config        *configv3.Config
	commandUI     *ui.UI
	plugin        configv3.Plugin
	commandParser rpc.CommandParser
}

// NewPluginRunner creates a new PluginRunner instance
func NewPluginRunner(config *configv3.Config, commandUI *ui.UI, plugin configv3.Plugin, commandParser rpc.CommandParser) PluginRunner {
	return &pluginRunner{
		config:        config,
		commandUI:     commandUI,
		plugin:        plugin,
		commandParser: commandParser,
	}
}

// Run executes the plugin with the given arguments
// Based on the plugin execution logic from cf/cmd/cmd.go (lines 140-163)
func (r *pluginRunner) Run(args []string) error {
	// Without a target the plugin still runs. RPC calls that need Cloud
	// Controller return rpc.ErrNoAPISet.
	var actor *v7action.Actor
	if r.commandUI != nil && r.config.Target() != "" {
		ccClient, uaaClient, routingClient, err := shared.GetNewClientsAndConnectToCF(r.config, r.commandUI, "")
		if err != nil {
			return fmt.Errorf("error connecting to CF: %w", err)
		}
		actor = v7action.NewActor(ccClient, r.config, sharedaction.NewActor(r.config), uaaClient, routingClient, clock.NewClock())
	}

	// Initialize RPC server
	server := netrpc.NewServer()
	rpcService, err := rpc.NewRpcService(
		r.config,
		server,
		actor,
		r.commandParser,
		r.commandUI,
	)
	if err != nil {
		return fmt.Errorf("error initializing RPC service: %w", err)
	}

	// Start RPC service
	err = rpcService.Start()
	if err != nil {
		return fmt.Errorf("error starting RPC service: %w", err)
	}
	// Like core commands, persist the config once the plugin is done, so tokens
	// refreshed by any RPC call survive. Deferred first, so it runs after Stop.
	defer func() {
		if writeErr := r.config.WriteConfig(); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %s", writeErr.Error())
		}
	}()
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

// CommandArgs returns the command line starting at the command: leading
// arguments starting with "-" are global options (which take no values) and are
// skipped. It returns nil when there is no command.
func CommandArgs(args []string) []string {
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return args[i:]
		}
	}
	return nil
}

// stopPlugin ensures the plugin process is terminated
func stopPlugin(plugin *exec.Cmd) {
	if plugin.Process != nil {
		// the plugin may already have exited; nothing to do if this fails
		_ = plugin.Process.Kill()
		_ = plugin.Wait()
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
