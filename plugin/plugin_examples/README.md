If you have any questions about developing a CLI plugin, ask away on the [cf-dev mailing list](https://lists.cloudfoundry.org/archives/list/cf-dev@lists.cloudfoundry.org/) (many plugin developers there!) or the #cli channel in our Slack community.

# Changes in v6.25.0
- `GetApp` now returns `Path` and `Port` information.

# Changes in v6.24.0
- API `LoggregatorEndpoint()` is deprecated and now always returns the empty string. Use `DopplerEndpoint()` instead to obtain logs.

# Changes in v6.17.0
- `-v` is now a global flag to enable verbose logging of API calls, equivalent to `CF_TRACE=true`. This means that the `-v` flag will no longer be passed to plugins.

# Changes in v6.14.0
- API `AccessToken()` now provides a refreshed o-auth token.
- [Examples](https://github.com/cloudfoundry/cli/tree/master/plugin/plugin_examples#test-driven-development-tdd) on how to use fake `CliConnection` and test RPC server for TDD development.
- Fix Plugin API file descriptors leakage.
- Fix bug where some CLI versions does not respect `PluginMetadata.MinCliVersion`.
- The field `PackageUpdatedAt` returned by `GetApp()` API is now populated.

[Complete change log ...](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/CHANGELOG.md)

# Developing a Plugin
[Go here for documentation of the plugin API](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/DOC.md)

This README discusses how to develop a cf CLI plugin.
For user-focused documentation, see [Using the cf CLI](http://docs.cloudfoundry.org/cf-cli/use-cli-plugins.html).

*If you wish to share your plugin with the community, see [here](http://github.com/cloudfoundry/cli-plugin-repo) for plugin submission.


## Development Requirements

- [GoLang installed](https://golang.org/doc/install)
- Tagged version of CLI release source code that supports plugins; cf CLI v.6.7.0 and above
```
mkdir -p "${GOPATH}/src/code.cloudfoundry.org"
cd "${GOPATH}/src/code.cloudfoundry.org"
git clone "https://github.com/cloudfoundry/cli"
```
(Optionally specify `--depth 1` to `git clone` for a faster download without any commit history)

## Architecture Overview

The cf CLI plugin architecture model follows the remote procedure call (RPC) model.
The cf CLI invokes each plugin, runs it as an independent executable, and handles all start, stop, and clean up tasks for plugin executable resources.

Here is an illustration of the work flow when a plugin command is being invoked.

1: CLI launches 2 processes, the rpc server and the independent plugin executable
<p align="center">
<img src="https://raw.githubusercontent.com/cloudfoundry/cli/master/plugin/plugin_examples/images/rpc_flow1.png" alt="workflow 1" width="400px">
</p>

2: Plugin establishes a connection to the RPC server, the connection is used to invoke core cli commands.
<p align="center">
<img src="https://raw.githubusercontent.com/cloudfoundry/cli/master/plugin/plugin_examples/images/rpc_flow2.png" alt="workflow 1" width="400px">
</p>

3: When a plugin invokes a cli command, it talks to the rpc server, and the rpc server interacts with cf cli to perform the command. The result is passed back to the plugin through the rpc server.
<p align="center">
<img src="https://raw.githubusercontent.com/cloudfoundry/cli/master/plugin/plugin_examples/images/rpc_flow3.png" alt="workflow 1" width="400px">
</p>

- Plugins that you develop for the cf CLI must conform to a predefined plugin interface that we discuss below.

## Writing a Plugin

[Go here for documentation of the plugin API](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/DOC.md)

To write a plugin for the cf CLI, implement the [predefined plugin interface](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin.go).

The interface uses a `Run(...)` method as the main entry point between the CLI and a plugin. This method receives the following arguments:

  - A struct `plugin.CliConnection` that contains methods for invoking cf CLI commands
  - A string array that contains the arguments passed from the `cf` process

The `GetMetadata()` function informs the CLI of the name of a plugin, plugin version (optional), minimum CLI version required (optional), the commands it implements, and help text for each command that users can display with `cf help`.

Plugin names with spaces must be enclosed in quotes when installed and uninstalled (e.g.: `cf install-plugin "my plugin"`). We recommend that plugin names not contain spaces to prevent the command shell from interpreting the name as multiple words.

  To initialize a plugin, call `plugin.Start(new(MyPluginStruct))` from within the `main()` method of your plugin. The `plugin.Start(...)` function requires a new reference to the struct that implements the defined interface.

This repo contains a basic plugin example [here](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/basic_plugin.go).<br>
To see more examples, go [here](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/).

### Uninstalling A Plugin
Uninstall of the plugin needs to be explicitly handled. When a user calls the `cf uninstall-plugin` command, CLI notifies the plugin via a call with `CLI-MESSAGE-UNINSTALL` as the first item in `[]args` from within the plugin's `Run(...)` method.

### Test Driven Development (TDD)
An example which was developed using TDD is available:
- `Test RPC server`: an RPC server to be used as a back-end for the plugin. It allows the plugin to be tested as a stand alone binary without replying on CLI as a back-end. [See example](https://github.com/cloudfoundry/cli/tree/master/plugin/plugin_examples/test_rpc_server_example)

### Using Command Line Arguments

The `Run(...)` method accepts the command line arguments and flags that you define for a plugin.

  See the [command line arguments example] (https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/echo.go) included in this repo.

#### Global Flags
There are several global flags that will not be passed to the plugin. These are:
- `-v`: equivalent to `CF_TRACE=true`, will display any API calls/responses to the user
- `-h`: will process the return from the plugin's `GetMetadata` function to produce a help display

### Calling CLI Commands

You can invoke CLI commands with `cliConnection.CliCommand([]args)` from within a plugin's `Run(...)` method. The `Run(...)` method receives the `cliConnection` as its first argument.

The `cliConnection.CliCommand([]args)` returns the output printed by the command and an error. The output is returned as a slice of strings. The error will be present if the call to the CLI command fails.

See the [test plugin example](https://github.com/cloudfoundry/cli/blob/master/integration/assets/test_plugin/test_plugin.go) included in this repo.

### Creating Interactive Plugins

Because a plugin has access to stdin during a call to the `Run(...)` method, you can create interactive plugins. See the [interactive plugin example](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/interactive.go) included in this repo.

### Creating Plugins with multiple commands

A single plugin binary can have more than one command, and each command can have it's own help text defined. For an example of multi-command plugins, see the [multiple commands example](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/multiple_commands.go)

### Enforcing a minimum CLI version required for the plugin.

```go
func (c *cmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Test1",
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 12,
			Build: 0,
		},
	}
}
```

### Debugging plugin code

The recommended approach to debugging plugin code is to print to stdout, or set CF_TRACE to /dev/stderr or a file.

## Compiling Plugin Source Code

The cf CLI requires an executable file to install the plugin. You must compile the source code with the `go build` command before distributing the plugin, or instruct your users to compile the plugin source code before installing the plugin. For information about compiling Go source code, see [Compile packages and dependencies](https://golang.org/cmd/go/).

## Using Plugins

After you compile a plugin, use the following commands to install and manage the plugin.

### Installing Plugins

To install a plugin, run:

`cf install-plugin PATH_TO_PLUGIN_BINARY`

### Listing Plugins

To display a list of installed plugins and the commands available from each plugin, run:

`cf plugins`

### Uninstalling Plugins

To remove a plugin, run:

`cf uninstall-plugin PLUGIN_NAME`

## Known Issues

- When invoking a CLI command using `cliConnection.CliCommand([]args)` a plugin will not receive output generated by the cli package. This includes usage failures when executing a cli command, `cf help`, or `cli SOME-COMMAND -h`.
- When invoking a CLI command using `cliConnection.CliCommand([]args)` and `CF_TRACE=true/cf -v` a plugin will receive all the output, including the trace in the returned string array. This may cause problem while trying to debug output with `CF_TRACE`. As work around, if a plugin is running `cf curl` via `CliCommand`, the following can be used to help with debugging (when the `CF_DEBUG_CURL=true`):
```go
  func RunCurl(cliConnection plugin.CliConnection, args []string) ([]string, error) {
    output, err := cliConnection.CliCommand("curl", args...)
    if os.Getenv("CF_DEBUG_CURL") == "true" {
      fmt.Println(strings.Join(output, "\n"))
    }
    return output, err
  }
```
- Due to architectural limitations, calling CLI core commands is not concurrency-safe. The correct execution of concurrent commands is not guaranteed. An architecture restructuring is in the works to fix this in the near future.
