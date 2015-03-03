package commands

import (
	"fmt"
	"sort"

	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/app_instances"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type AppManifestCreater interface {
	CreateAppManifest(app models.Application, orgName string, spaceName string)
}

type CreateAppManifest struct {
	ui               terminal.UI
	config           core_config.Reader
	appSummaryRepo   api.AppSummaryRepository
	appInstancesRepo app_instances.AppInstancesRepository
	appReq           requirements.ApplicationRequirement
	manifest         manifest.AppManifest
}

func NewCreateAppManifest(ui terminal.UI, config core_config.Reader, appSummaryRepo api.AppSummaryRepository, manifestGenerator manifest.AppManifest) (cmd *CreateAppManifest) {
	cmd = new(CreateAppManifest)
	cmd.ui = ui
	cmd.config = config
	cmd.appSummaryRepo = appSummaryRepo
	cmd.manifest = manifestGenerator
	return
}

func (cmd *CreateAppManifest) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-app-manifest",
		Description: T("Create an app manifest for an app that has been pushed successfully."),
		Usage:       T("CF_NAME create-app-manifest APP [-p /path/to/<app-name>-manifest.yml ]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", T("Specify a path for file creation. If path not specified, manifest file is created in current working directory.")),
		},
	}
}

func (cmd *CreateAppManifest) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	if cmd.appReq == nil {
		cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])
	} else {
		cmd.appReq.SetApplicationName(c.Args()[0])
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *CreateAppManifest) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	application, apiErr := cmd.appSummaryRepo.GetSummary(app.Guid)

	if apiErr != nil {
		cmd.ui.Failed(T("Error getting application summary: ") + apiErr.Error())
	}

	cmd.ui.Say(T("Creating an app manifest from current settings of app ") + application.Name + " ...")
	cmd.ui.Say("")

	savePath := "./" + application.Name + "_manifest.yml"

	if c.String("p") != "" {
		savePath = c.String("p")
	}

	cmd.createManifest(application, savePath)
}

func (cmd *CreateAppManifest) createManifest(app models.Application, savePath string) error {
	cmd.manifest.FileSavePath(savePath)
	cmd.manifest.Memory(app.Name, app.Memory)
	cmd.manifest.Instances(app.Name, app.InstanceCount)

	if app.Command != "" {
		cmd.manifest.StartupCommand(app.Name, app.Command)
	}

	if len(app.Services) > 0 {
		for _, service := range app.Services {
			cmd.manifest.Service(app.Name, service.Name)
		}
	}

	if app.HealthCheckTimeout > 0 {
		cmd.manifest.HealthCheckTimeout(app.Name, app.HealthCheckTimeout)
	}

	if len(app.EnvironmentVars) > 0 {
		sorted := sortEnvVar(app.EnvironmentVars)
		for _, envVarKey := range sorted {
			switch app.EnvironmentVars[envVarKey].(type) {
			default:
				cmd.ui.Failed(T("Failed to create manifest, unable to parse environment variable: ") + envVarKey)
			case float64:
				//json.Unmarshal turn all numbers to float64
				value := int(app.EnvironmentVars[envVarKey].(float64))
				cmd.manifest.EnvironmentVars(app.Name, envVarKey, fmt.Sprintf("%d", value))
			case bool:
				cmd.manifest.EnvironmentVars(app.Name, envVarKey, fmt.Sprintf("%t", app.EnvironmentVars[envVarKey].(bool)))
			case string:
				cmd.manifest.EnvironmentVars(app.Name, envVarKey, "\""+app.EnvironmentVars[envVarKey].(string)+"\"")
			}
		}
	}

	if len(app.Routes) > 0 {
		cmd.manifest.Domain(app.Name, app.Routes[0].Host, app.Routes[0].Domain.Name)
	}

	err := cmd.manifest.Save()
	if err != nil {
		cmd.ui.Failed(T("Error creating manifest file: ") + err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("Manifest file created successfully at ") + savePath)
	cmd.ui.Say("")

	return nil
}

func sortEnvVar(vars map[string]interface{}) []string {
	var varsAry []string
	for k, _ := range vars {
		varsAry = append(varsAry, k)
	}
	sort.Strings(varsAry)

	return varsAry
}
