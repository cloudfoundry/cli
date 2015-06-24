package command_registry

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/i18n/detection"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/plugin/models"
)

type Dependency struct {
	Ui           terminal.UI
	Config       core_config.Repository
	RepoLocator  api.RepositoryLocator
	Detector     detection.Detector
	PluginConfig plugin_config.PluginConfiguration
	ManifestRepo manifest.ManifestRepository
	Gateways     map[string]net.Gateway
	TeePrinter   *terminal.TeePrinter
	PluginModels *pluginModels
}

type pluginModels struct {
	Application      *plugin_models.GetAppModel
	AppsSummary      *[]plugin_models.GetAppsModel
	Organizations    *[]plugin_models.OrganizationSummary
	Organization     *plugin_models.Organization
	Spaces           *[]plugin_models.SpaceSummary
	Space            *plugin_models.Space
	Users            *[]plugin_models.User
	ServiceInstances *[]plugin_models.ServiceInstance
}

func NewDependency() Dependency {
	deps := Dependency{}
	deps.TeePrinter = terminal.NewTeePrinter()
	deps.Ui = terminal.NewUI(os.Stdin, deps.TeePrinter)
	deps.ManifestRepo = manifest.NewManifestDiskRepository()

	errorHandler := func(err error) {
		if err != nil {
			deps.Ui.Failed(fmt.Sprintf("Config error: %s", err))
		}
	}
	deps.Config = core_config.NewRepositoryFromFilepath(config_helpers.DefaultFilePath(), errorHandler)
	deps.PluginConfig = plugin_config.NewPluginConfig(errorHandler)
	deps.Detector = &detection.JibberJabberDetector{}

	terminal.UserAskedForColors = deps.Config.ColorEnabled()
	terminal.InitColorSupport()

	if os.Getenv("CF_TRACE") != "" {
		trace.Logger = trace.NewLogger(os.Getenv("CF_TRACE"))
	} else {
		trace.Logger = trace.NewLogger(deps.Config.Trace())
	}

	deps.Gateways = map[string]net.Gateway{
		"auth":             net.NewUAAGateway(deps.Config, deps.Ui),
		"cloud-controller": net.NewCloudControllerGateway(deps.Config, time.Now, deps.Ui),
		"uaa":              net.NewUAAGateway(deps.Config, deps.Ui),
	}
	deps.RepoLocator = api.NewRepositoryLocator(deps.Config, deps.Gateways)

	deps.PluginModels = &pluginModels{Application: nil}

	return deps
}
