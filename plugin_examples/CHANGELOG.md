# Changes in v6.11.0
-Plugins now have a hook-in that is called when the plugin is uninstalled, allowing cleanup of files.

# Changes in v6.10.0
[CF-Community Plugin Repository](https://github.com/cloudfoundry-incubator/cli-plugin-repo) introduced.
- Plugin developers can submit any open-source plugins 
- Plugins in the community repo can be browsed and installed from the CLI

# Changes in v6.9.0
- Plugins can now have versions, i.e. 1.2.3, [code example](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/basic_plugin.go)
- `cf plugins` now displays plugin versions
- `-h` and `--help` flags work with plugin commands. e.g. `cf <plugin-command> -h`. [code example](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/echo.go)
- Allow `cf help <plugin-command>`

# Changes in v6.8.0
- Plugin commands can now have aliases
- Help text for plugins now listed in 'cf plugins'

# Changes in v6.7.0
- Plugins introduced
