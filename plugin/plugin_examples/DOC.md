
## Plugin API
We wrote the Plugin API to make it easy for plugins to consume output from calling CLI commands.  Previously, plugins needed to parse the terminal output which was not optimal.  Before we wrote the API, only 2 methods were available to plugins: 
```
CliCommand()
CliCommandWithoutTerminalOutput() 
``` 

Both commands returned the terminal output in a string array, which was hard to parse. Instead terminal output, the result of the API calls will be in an object which is much easier to parse.  Our goal  was to make the common resources readily available to plugins without parsing.




Latest Available API Commands
```go

/******************************************************************
returns the output printed by the command and an error.
The output is returned as a slice of strings.
The error will be present if the call to the CLI command fails.
******************************************************************/
CliCommand(args ...string) ([]string, error)

/******************************************************************
  just like CliCommand but without the output in the terminal
******************************************************************/  
CliCommandWithoutTerminalOutput(args ...string) ([]string, error)

GetCurrentOrg() (plugin_models.Organization, error)

GetCurrentSpace() (plugin_models.Space, error)

Username() (userName string, error)

UserGuid() (userGuid string, error)

UserEmail() (userEmail string, error)

IsLoggedIn() (bool, error)

IsSSLDisabled() (bool, error)

HasOrganization() (bool, error)

HasSpace() (bool, error)

ApiEndpoint() (endpointUrl string, error)

ApiVersion() (ver string, error)

HasAPIEndpoint() (bool, error)

LoggregatorEndpoint() (endpointUrl string, error)

DopplerEndpoint() (endpointUrl string, error)

AccessToken() (token string, error)

GetApp(string) (plugin_models.GetAppModel, error)

GetApps() ([]plugin_models.GetAppsModel, error)

GetOrgs() ([]plugin_models.GetOrgs_Model, error)

GetOrg(string) (plugin_models.GetOrg_Model, error)

GetSpaces() ([]plugin_models.GetSpaces_Model, error)

GetSpace(spaceName string) (plugin_models.GetSpace_Model, error)

/******************************************************************
options takes the optional argument used in the `cf org` command, see `cf org -h`
******************************************************************/
GetOrgUsers(orgName string, options ...string) ([]plugin_models.GetOrgUsers_Model, error)

GetSpaceUsers(orgName string, spaceName string) ([]plugin_models.GetSpaceUsers_Model, error)

GetServices() ([]plugin_models.GetServices_Model, error)

GetService(serviceInstance string) (plugin_models.GetService_Model, error)
```
---
Models return from APIs
- [Organization](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_current_org.go#L3)
- [Space](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_current_space.go#L3)
- [GetApp_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_app.go#L5)
- [GetApps_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_apps.go#L3)
- [GetOrgs_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_orgs.go#L3)
- [GetOrg_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_org.go#L3)
- [GetSpaces_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_spaces.go#L3)
- [GetSpace_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_space.go#L3)
- [GetOrgUsers_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_org_users.go#L3)
- [GetSpaceUsers_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_space_users.go#L3)
- [GetServices_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_services.go#L3)
- [GetService_Model](https://github.com/cloudfoundry/cli/blob/master/plugin/models/get_service.go#L3)
