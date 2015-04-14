# Changes in v6.11.0
-Plugins now have a hook-in that is called when the plugin is uninstalled, allowing cleanup of files.

[Complete change log ...](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/CHANGELOG.md) 

# Developing a Plugin

This README discusses how to develop a cf CLI plugin.
For user-focused documentation, see [Using the cf CLI](http://docs.cloudfoundry.org/devguide/installcf/use-cli-plugins.html).

*If you wish to share your plugin with the community, see [here](http://github.com/cloudfoundry-incubator/cli-plugin-repo) for plugin submission.


## Development Requirements

- Golang installed
- Tagged version of CLI release source code that supports plugins; cf CLI v.6.7.0 and above

## Architecture Overview

The cf CLI plugin architecture model follows the remote procedure call (RPC) model.
The cf CLI invokes each plugin, runs it as an independent executable, and handles all start, stop, and clean up tasks for plugin executable resources.

Here is an illustration of the workflow when a plugin command is being invoked.

1: CLI launches 2 processes, the rpc server and the independent plugin executable
<p align="center">
<img src="https://raw.githubusercontent.com/cloudfoundry/cli/master/plugin_examples/images/rpc_flow1.png" alt="workflow 1" width="400px">
</p>

2: Plugin establishes a connection to the RPC server, the connection is used to invoke core cli commands.
<p align="center">
<img src="https://raw.githubusercontent.com/cloudfoundry/cli/master/plugin_examples/images/rpc_flow2.png" alt="workflow 1" width="400px">
</p>

3: When a plugin invokes a cli command, it talks to the rpc server, and the rpc server interacts with cf cli to perform the command. The result is passed back to the plugin through the rpc server.
<p align="center">
<img src="https://raw.githubusercontent.com/cloudfoundry/cli/master/plugin_examples/images/rpc_flow3.png" alt="workflow 1" width="400px">
</p>

- Plugins that you develop for the cf CLI must conform to a predefined plugin interface that we discuss below.

## Writing a Plugin

To write a plugin for the cf CLI, implement the 
[predefined plugin interface](https://github.com/cloudfoundry/cli/blob/master/plugin/plugin.go).

The interface uses a `Run(...)` method as the main entry point between the CLI 
and a plugin. This method receives the following arguments:

  - A struct `plugin.CliConnection` that contains methods for invoking cf CLI commands
  - A string array that contains the arguments passed from the `cf` process

The `GetMetadata()` function informs the CLI of the name of a plugin, plugin version (optional), the 
commands it implements, and help text for each command that users can display 
with `cf help`.

  To initialize a plugin, call `plugin.Start(new(MyPluginStruct))` from within the `main()` method of your plugin. The `plugin.Start(...)` function requires a new reference to the struct that implements the defined interface. 

This repo contains a basic plugin example [here](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/basic_plugin.go).<br>
To see more examples, go [here](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/).

If you wish to employ TDD in your plugin development, [here](https://github.com/cloudfoundry/cli/tree/master/plugin_examples/call_cli_cmd/main) is an example of a plugin that calls core cli command with the use of `FakeCliConnection` for testing.

### Using Command Line Arguments

The `Run(...)` method accepts the command line arguments and flags that you 
define for a plugin. 

  See the [command line arguments example] (https://github.com/cloudfoundry/cli/blob/master/plugin_examples/echo.go) included in this repo.

### Calling CLI Commands

You can invoke CLI commands with `cliConnection.CliCommand([]args)` from
 within a plugin's `Run(...)` method. The `Run(...)` method receives the 
`cliConnection` as its first argument.

The `cliConnection.CliCommand([]args)` returns the output printed by the command and an error. The output is returned as a slice of strings. The error 
will be present if the call to the CLI command fails.

See the [calling CLI commands example](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/call_cli_cmd/main/call_cli_cmd.go) included in this repo.

### Creating Interactive Plugins

Because a plugin has access to stdin during a call to the `Run(...)` method, you can create interactive plugins. See the [interactive plugin example](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/interactive.go)
 included in this repo. 

### Creating Plugins with multiple commands

A single plugin binary can have more than one command, and each command can have it's own help text defined. For an example of multi-comamnd plugins, see the [multiple commands example](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/multiple_commands.go)

## Compiling Plugin Source Code

The cf CLI requires an executable file to install the plugin. You must compile the source code with the `go build` command before distributing the plugin, or instruct your users to compile the plugin source code before 
installing the plugin. For information about compiling Go source code, see [Compile packages and dependencies](https://golang.org/cmd/go/).  

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

- When invoking a CLI command using `cliConnection.CliCommand([]args)` a plugin developer will not receive output generated by the codegangsta/cli package. This includes usage failures when executing a cli command, `cf help`, or `cli SOME-COMMAND -h`. 
- Due to architectural limitations, calling CLI core commands is not concurrency-safe. The correct execution of concurrent commands is not guaranteed. An architecture restructuring is in the works to fix this in the near future.
   

