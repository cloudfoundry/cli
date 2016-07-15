package commandregistry

import (
	"fmt"
	"io"
	"os"
	"time"

	"path/filepath"

	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/actors/brokerbuilder"
	"github.com/cloudfoundry/cli/cf/actors/planbuilder"
	"github.com/cloudfoundry/cli/cf/actors/pluginrepo"
	"github.com/cloudfoundry/cli/cf/actors/servicebuilder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/appfiles"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/configuration/confighelpers"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/plugin/models"
	"github.com/cloudfoundry/cli/utils"
	"github.com/cloudfoundry/cli/utils/words/generator"
)

type Dependency struct {
	UI                 terminal.UI
	Config             coreconfig.Repository
	RepoLocator        api.RepositoryLocator
	PluginConfig       pluginconfig.PluginConfiguration
	ManifestRepo       manifest.Repository
	AppManifest        manifest.App
	Gateways           map[string]net.Gateway
	TeePrinter         *terminal.TeePrinter
	PluginRepo         pluginrepo.PluginRepo
	PluginModels       *PluginModels
	ServiceBuilder     servicebuilder.ServiceBuilder
	BrokerBuilder      brokerbuilder.Builder
	PlanBuilder        planbuilder.PlanBuilder
	ServiceHandler     actors.ServiceActor
	ServicePlanHandler actors.ServicePlanActor
	WordGenerator      generator.WordGenerator
	AppZipper          appfiles.Zipper
	AppFiles           appfiles.AppFiles
	PushActor          actors.PushActor
	RouteActor         actors.RouteActor
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

func NewDependency(writer io.Writer, logger trace.Printer, envDialTimeout string) Dependency {
	deps := Dependency{}
	deps.TeePrinter = terminal.NewTeePrinter(writer)
	deps.UI = terminal.NewUI(os.Stdin, writer, deps.TeePrinter, logger)

	errorHandler := func(err error) {
		if err != nil {
			deps.UI.Failed(fmt.Sprintf("Config error: %s", err))
		}
	}

	configPath, err := confighelpers.DefaultFilePath()
	if err != nil {
		errorHandler(err)
	}
	deps.Config = coreconfig.NewRepositoryFromFilepath(configPath, errorHandler)

	deps.ManifestRepo = manifest.NewDiskRepository()
	deps.AppManifest = manifest.NewGenerator()

	pluginPath := filepath.Join(confighelpers.PluginRepoDir(), ".cf", "plugins")
	deps.PluginConfig = pluginconfig.NewPluginConfig(
		errorHandler,
		configuration.NewDiskPersistor(filepath.Join(pluginPath, "config.json")),
		pluginPath,
	)

	terminal.UserAskedForColors = deps.Config.ColorEnabled()
	terminal.InitColorSupport()

	deps.Gateways = map[string]net.Gateway{
		"cloud-controller": net.NewCloudControllerGateway(deps.Config, time.Now, deps.UI, logger, envDialTimeout),
		"uaa":              net.NewUAAGateway(deps.Config, deps.UI, logger, envDialTimeout),
		"routing-api":      net.NewRoutingAPIGateway(deps.Config, time.Now, deps.UI, logger, envDialTimeout),
	}
	deps.RepoLocator = api.NewRepositoryLocator(deps.Config, deps.Gateways, logger)

	deps.PluginModels = &PluginModels{Application: nil}

	deps.PlanBuilder = planbuilder.NewBuilder(
		deps.RepoLocator.GetServicePlanRepository(),
		deps.RepoLocator.GetServicePlanVisibilityRepository(),
		deps.RepoLocator.GetOrganizationRepository(),
	)

	deps.ServiceBuilder = servicebuilder.NewBuilder(
		deps.RepoLocator.GetServiceRepository(),
		deps.PlanBuilder,
	)

	deps.BrokerBuilder = brokerbuilder.NewBuilder(
		deps.RepoLocator.GetServiceBrokerRepository(),
		deps.ServiceBuilder,
	)

	deps.PluginRepo = pluginrepo.NewPluginRepo()

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

	deps.AppZipper = appfiles.ApplicationZipper{}
	deps.AppFiles = appfiles.ApplicationFiles{}

	deps.RouteActor = actors.NewRouteActor(deps.UI, deps.RepoLocator.GetRouteRepository(), deps.RepoLocator.GetDomainRepository())
	deps.PushActor = actors.NewPushActor(deps.RepoLocator.GetApplicationBitsRepository(), deps.AppZipper, deps.AppFiles, deps.RouteActor)

	deps.ChecksumUtil = utils.NewSha1Checksum("")

	deps.Logger = logger

	return deps
}
