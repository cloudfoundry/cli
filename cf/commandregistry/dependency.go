package commandregistry

import (
	"fmt"
	"io"
	"os"
	"time"

	"path/filepath"

	"code.cloudfoundry.org/cli/cf/actors"
	"code.cloudfoundry.org/cli/cf/actors/brokerbuilder"
	"code.cloudfoundry.org/cli/cf/actors/planbuilder"
	"code.cloudfoundry.org/cli/cf/actors/pluginrepo"
	"code.cloudfoundry.org/cli/cf/actors/servicebuilder"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/appfiles"
	"code.cloudfoundry.org/cli/cf/configuration"
	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
	"code.cloudfoundry.org/cli/cf/manifest"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/util"
	"code.cloudfoundry.org/cli/util/randomword"
)

//go:generate counterfeiter . RandomWordGenerator

type RandomWordGenerator interface {
	Babble() string
}

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
	WordGenerator      RandomWordGenerator
	AppZipper          appfiles.Zipper
	AppFiles           appfiles.AppFiles
	PushActor          actors.PushActor
	RouteActor         actors.RouteActor
	ChecksumUtil       util.Sha1Checksum
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
	deps.RepoLocator = api.NewRepositoryLocator(deps.Config, deps.Gateways, logger, envDialTimeout)

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

	deps.WordGenerator = new(randomword.Generator)

	deps.AppZipper = appfiles.ApplicationZipper{}
	deps.AppFiles = appfiles.ApplicationFiles{}

	deps.RouteActor = actors.NewRouteActor(deps.UI, deps.RepoLocator.GetRouteRepository(), deps.RepoLocator.GetDomainRepository())
	deps.PushActor = actors.NewPushActor(deps.RepoLocator.GetApplicationBitsRepository(), deps.AppZipper, deps.AppFiles, deps.RouteActor)

	deps.ChecksumUtil = util.NewSha1Checksum("")

	deps.Logger = logger

	return deps
}
