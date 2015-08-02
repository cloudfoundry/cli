[Go here for documentation of the plugin API](https://github.com/cloudfoundry/cli/blob/master/plugin_examples/DOC.md)

# Changes in v6.12.0
- New API:
```go
GetApp(string) (plugin_models.GetAppModel, error)
GetApps() ([]plugin_models.GetAppsModel, error)
GetOrgs() ([]plugin_models.GetOrgs_Model, error)
GetSpaces() ([]plugin_models.GetSpaces_Model, error)
GetOrgUsers(string, ...string) ([]plugin_models.GetOrgUsers_Model, error)
GetSpaceUsers(string, string) ([]plugin_models.GetSpaceUsers_Model, error)
GetServices() ([]plugin_models.GetServices_Model, error)
GetService(string) (plugin_models.GetService_Model, error)
GetOrg(string) (plugin_models.GetOrg_Model, error)
GetSpace(string) (plugin_models.GetSpace_Model, error)
```
- Allow minimum CLI version required to be specified in plugin. Example:
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

# Changes in v6.11.2
Added the following commands to cli_connection.go:
```go
  - GetCurrentOrg()  
  - GetCurrentSpace()  
  - Username()  
  - UserEmail()  
  - UserGuid()  
  - HasOrganization()  
  - HasSpace()  
  - IsLoggedIn()  
  - IsSSLDisabled()  
  - ApiEndpoint()  
  - HasAPIEndpoint()  
  - ApiVersion()  
  - LoggregatorEndpoint()  
  - DopplerEndpoint()  
  - AccessToken()  
```

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
