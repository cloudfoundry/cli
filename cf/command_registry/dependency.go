package command_registry

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/actors/brokerbuilder"
	"github.com/cloudfoundry/cli/cf/actors/planbuilder"
	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/configuration/confighelpers"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/plugin/models"
	"github.com/cloudfoundry/cli/utils"
	"github.com/cloudfoundry/cli/words/generator"
)

type Dependency struct {
	Ui                 terminal.UI
	Config             coreconfig.Repository
	RepoLocator        api.RepositoryLocator
	PluginConfig       pluginconfig.PluginConfiguration
	ManifestRepo       manifest.ManifestRepository
	AppManifest        manifest.AppManifest
	Gateways           map[string]net.Gateway
	TeePrinter         *terminal.TeePrinter
	PluginRepo         plugin_repo.PluginRepo
	PluginModels       *PluginModels
	ServiceBuilder     service_builder.ServiceBuilder
	BrokerBuilder      brokerbuilder.Builder
	PlanBuilder        planbuilder.PlanBuilder
	ServiceHandler     actors.ServiceActor
	ServicePlanHandler actors.ServicePlanActor
	WordGenerator      generator.WordGenerator
	AppZipper          app_files.Zipper
	AppFiles           app_files.AppFiles
	PushActor          actors.PushActor
	ChecksumUtil       utils.Sha1Checksum
	WildcardDependency interface{} //use for injecting fakes
	Logger             trace.Printer
}

type PluginModels struct {
	Application   *plugin_models.GetAppModel
	AppsSummary   *[]plugin_models.GetAppsModel
	Organizations *[]plugin_models.GetOrgs_Model
	Organization  *plugin_models.GetOrg_Model
	Spaces        *[]plugin_models.GetSpaces_Model
	Space         *plugin_models.GetSpace_Model
	OrgUsers      *[]plugin_models.GetOrgUsers_Model
	SpaceUsers    *[]plugin_models.GetSpaceUsers_Model
	Services      *[]plugin_models.GetServices_Model
	Service       *plugin_models.GetService_Model
	OauthToken    *plugin_models.GetOauthToken_Model
}

func NewDependency(logger trace.Printer) Dependency {
	deps := Dependency{}
	deps.TeePrinter = terminal.NewTeePrinter()
	deps.Ui = terminal.NewUI(os.Stdin, deps.TeePrinter, logger)

	errorHandler := func(err error) {
		if err != nil {
			deps.Ui.Failed(fmt.Sprintf("Config error: %s", err))
		}
	}
	deps.Config = coreconfig.NewRepositoryFromFilepath(confighelpers.DefaultFilePath(), errorHandler)

	deps.ManifestRepo = manifest.NewManifestDiskRepository()
	deps.AppManifest = manifest.NewGenerator()
	deps.PluginConfig = pluginconfig.NewPluginConfig(errorHandler)

	terminal.UserAskedForColors = deps.Config.ColorEnabled()
	terminal.InitColorSupport()

	deps.Gateways = map[string]net.Gateway{
		"cloud-controller": net.NewCloudControllerGateway(deps.Config, time.Now, deps.Ui, logger),
		"uaa":              net.NewUAAGateway(deps.Config, deps.Ui, logger),
		"routing-api":      net.NewRoutingApiGateway(deps.Config, time.Now, deps.Ui, logger),
	}
	deps.RepoLocator = api.NewRepositoryLocator(deps.Config, deps.Gateways, logger)

	deps.PluginModels = &PluginModels{Application: nil}

	deps.PlanBuilder = planbuilder.NewBuilder(
		deps.RepoLocator.GetServicePlanRepository(),
		deps.RepoLocator.GetServicePlanVisibilityRepository(),
		deps.RepoLocator.GetOrganizationRepository(),
	)

	deps.ServiceBuilder = service_builder.NewBuilder(
		deps.RepoLocator.GetServiceRepository(),
		deps.PlanBuilder,
	)

	deps.BrokerBuilder = brokerbuilder.NewBuilder(
		deps.RepoLocator.GetServiceBrokerRepository(),
		deps.ServiceBuilder,
	)

	deps.PluginRepo = plugin_repo.NewPluginRepo()

	deps.ServiceHandler = actors.NewServiceHandler(
		deps.RepoLocator.GetOrganizationRepository(),
		deps.BrokerBuilder,
		deps.ServiceBuilder,
	)

	deps.ServicePlanHandler = actors.NewServicePlanHandler(
		deps.RepoLocator.GetServicePlanRepository(),
		deps.RepoLocator.GetServicePlanVisibilityRepository(),
		deps.RepoLocator.GetOrganizationRepository(),
		deps.PlanBuilder,
		deps.ServiceBuilder,
	)

	deps.WordGenerator = generator.NewWordGenerator()

	deps.AppZipper = app_files.ApplicationZipper{}
	deps.AppFiles = app_files.ApplicationFiles{}

	deps.PushActor = actors.NewPushActor(deps.RepoLocator.GetApplicationBitsRepository(), deps.AppZipper, deps.AppFiles)

	deps.ChecksumUtil = utils.NewSha1Checksum("")

	deps.Logger = logger

	return deps
}
