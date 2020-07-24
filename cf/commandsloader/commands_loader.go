package commandsloader

import (
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/commands/plugin"
	"code.cloudfoundry.org/cli/cf/commands/pluginrepo"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/commands/serviceaccess"
	"code.cloudfoundry.org/cli/cf/commands/serviceauthtoken"
	"code.cloudfoundry.org/cli/cf/commands/servicebroker"
	"code.cloudfoundry.org/cli/cf/commands/servicekey"
)

/*******************
This package make a reference to all the command packages
in cf/commands/..., so all init() in the directories will
get initialized

* Any new command packages must be included here for init() to get called
********************/

func Load() {
	_ = commands.API{}
	_ = plugin.Plugins{}
	_ = pluginrepo.RepoPlugins{}
	_ = service.ShowService{}
	_ = serviceauthtoken.ListServiceAuthTokens{}
	_ = serviceaccess.ServiceAccess{}
	_ = servicebroker.ListServiceBrokers{}
	_ = servicekey.ServiceKey{}
}
