# Developing a Plugin

## Development Requirements

  - Golang installed
  - Tagged version of CLI release source code supporting plugins

## Architecture Overview

  The CLI plugin architecture is built on an rpc model. This means that each plugin
  is run as an independent executable and is invoked by the CLI. The CLI
  handles starting, stopping, and cleaning up the plugin executable resources.
  To write a plugin compatible with the CLI a developer only has to conform to
  a simple interface defined [here](github.com/cloudfoundry/cli/plugin/plugin.go)

## Writing Your First Plugin

  To start writting a plugin for the CF CLI, a developer will need to implement
  a predefined `Plugin` interface which can be found [here](github.com/cloudfoundry/cli/plugin/plugin.go)

  The `Run(...)` method is used at the main entry point between the CLI
  and a plugin. The run method receives two arguments. The first argument is
  a plugin.CliConnection. The plugin.CliConnection is a struct containing methods 
  for invoking cf cli commands. The second argument to Run(..) is a slice 
  containing the arguments passed from the `cf` process.

  The `GetMetadata()` function informs the CLI of a plugin name, the
  commands it implements and help text for each command to be displayed with
  `cf help`.

  Initializing a plugin is as easy as calling 
  `plugin.Start(new(MyPluginStruct))`. The function `plugin.Start(...)`
  requires a new reference to the struct implementing the defined interface.
  The `plugin.Start(...)` method should be invoked in a plugin's `main()`
  method.

  A basic plugin example can be found [here](github.com/cloudfoundry/plugin_examples/basic_plugin.go)

  Plugins need to be compiled before installation. Information about
  building a binary can be found [here](https://www.google.com/search?q=how%20to%20compile%20golang)

## Installing Plugins

  A compiled plugin can be installed by invoking

  `cf install-plugin path/to/plugin-binary`

## Listing Plugins

  To see a list of installed plugins run

  `cf plugins`

  This shows a list of the commands that are avaiable from installed plugins,
  along with each command's plugin name.

## Uninstalling Plugins

  A plugin is uninstalled by running the command

  `cf uninstall-plugin <plugin-name>`

## Command Line Arguments

  Command line arguments are sent along to plugins via the `Run(...)` method.

  An example plugin that parses command line arguments and flags can be
  found [here](github.com/cloudfoundry/plugin_examples/echo.go).

## Calling CLI Commands

  CLI commands can be invoked with `cliConnection.CliCommand([]args)` from
  within a plugin's `Run(..)` method. The Run(..) method receives the 
  cliConnection as the first argument to Run(..)

  The `plugin.CliCommand([]args)` returns the output printed by the command
  and an error. The output is returned as a slice of strings. The error
  will be present if the call to the cli command fails.

  An example usage can be found [here](github.com/cloudfoundry/plugin_examples/call_cli_cmd/main/call_cli_cmd.go)

## Interactive Plugins

  Plugins have the ability to be interactive. During a call to `Run(...)` a
  plugin has access to stdin.

  An example usage can be found [here](github.com/cloudfoundry/plugin_examples/interactive.go)
