package commands_loader

import (
	"github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/commands/buildpack"
	"github.com/cloudfoundry/cli/cf/commands/domain"
	"github.com/cloudfoundry/cli/cf/commands/environmentvariablegroup"
	"github.com/cloudfoundry/cli/cf/commands/featureflag"
	"github.com/cloudfoundry/cli/cf/commands/organization"
	"github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/cf/commands/plugin_repo"
	"github.com/cloudfoundry/cli/cf/commands/quota"
	"github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/commands/securitygroup"
	"github.com/cloudfoundry/cli/cf/commands/serviceaccess"
	"github.com/cloudfoundry/cli/cf/commands/serviceauthtoken"
	"github.com/cloudfoundry/cli/cf/commands/servicebroker"
	"github.com/cloudfoundry/cli/cf/commands/servicekey"
	"github.com/cloudfoundry/cli/cf/commands/space"
	"github.com/cloudfoundry/cli/cf/commands/spacequota"
)

/*******************
This package make a reference to all the command packages
in cf/commands/..., so all init() in the directories will
get initialized

* Any new command packages must be included here for init() to get called
********************/

func Load() {
	_ = application.ListApps{}
	_ = domain.CreateDomain{}
	_ = buildpack.ListBuildpacks{}
	_ = quota.CreateQuota{}
	_ = organization.ListOrgs{}
	_ = spacequota.SpaceQuota{}
	_ = servicebroker.ListServiceBrokers{}
	_ = serviceauthtoken.ListServiceAuthTokens{}
	_ = securitygroup.ShowSecurityGroup{}
	_ = environmentvariablegroup.RunningEnvironmentVariableGroup{}
	_ = featureflag.ShowFeatureFlag{}
	_ = commands.Api{}
	_ = plugin_repo.RepoPlugins{}
	_ = plugin.Plugins{}
	_ = route.CreateRoute{}
	_ = space.CreateSpace{}
	_ = serviceaccess.ServiceAccess{}
	_ = servicekey.ServiceKey{}
}
