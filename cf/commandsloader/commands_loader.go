package commandsloader

import (
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/commands/application"
	"code.cloudfoundry.org/cli/cf/commands/buildpack"
	"code.cloudfoundry.org/cli/cf/commands/domain"
	"code.cloudfoundry.org/cli/cf/commands/environmentvariablegroup"
	"code.cloudfoundry.org/cli/cf/commands/featureflag"
	"code.cloudfoundry.org/cli/cf/commands/organization"
	"code.cloudfoundry.org/cli/cf/commands/plugin"
	"code.cloudfoundry.org/cli/cf/commands/pluginrepo"
	"code.cloudfoundry.org/cli/cf/commands/quota"
	"code.cloudfoundry.org/cli/cf/commands/route"
	"code.cloudfoundry.org/cli/cf/commands/routergroups"
	"code.cloudfoundry.org/cli/cf/commands/securitygroup"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/commands/serviceaccess"
	"code.cloudfoundry.org/cli/cf/commands/serviceauthtoken"
	"code.cloudfoundry.org/cli/cf/commands/servicebroker"
	"code.cloudfoundry.org/cli/cf/commands/servicekey"
	"code.cloudfoundry.org/cli/cf/commands/space"
	"code.cloudfoundry.org/cli/cf/commands/spacequota"
	"code.cloudfoundry.org/cli/cf/commands/user"
)

/*******************
This package make a reference to all the command packages
in cf/commands/..., so all init() in the directories will
get initialized

* Any new command packages must be included here for init() to get called
********************/

func Load() {
	_ = commands.API{}
	_ = application.ListApps{}
	_ = buildpack.ListBuildpacks{}
	_ = domain.CreateDomain{}
	_ = environmentvariablegroup.RunningEnvironmentVariableGroup{}
	_ = featureflag.ShowFeatureFlag{}
	_ = organization.ListOrgs{}
	_ = plugin.Plugins{}
	_ = pluginrepo.RepoPlugins{}
	_ = quota.CreateQuota{}
	_ = route.CreateRoute{}
	_ = routergroups.RouterGroups{}
	_ = securitygroup.ShowSecurityGroup{}
	_ = service.ShowService{}
	_ = serviceauthtoken.ListServiceAuthTokens{}
	_ = serviceaccess.ServiceAccess{}
	_ = servicebroker.ListServiceBrokers{}
	_ = servicekey.ServiceKey{}
	_ = space.CreateSpace{}
	_ = spacequota.SpaceQuota{}
	_ = user.CreateUser{}
}
