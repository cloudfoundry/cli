package commands

import (
	"fmt"
	"os"
	"sort"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/app_instances"
	"github.com/cloudfoundry/cli/cf/api/stacks"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type CreateAppManifest struct {
	ui               terminal.UI
	config           core_config.Reader
	appSummaryRepo   api.AppSummaryRepository
	stackRepo        stacks.StackRepository
	appInstancesRepo app_instances.AppInstancesRepository
	appReq           requirements.ApplicationRequirement
	manifest         manifest.AppManifest
}

func init() {
	command_registry.Register(&CreateAppManifest{})
}

func (cmd *CreateAppManifest) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Specify a path for file creation. If path not specified, manifest file is created in current working directory.")}

	return command_registry.CommandMetadata{
		Name:        "create-app-manifest",
		Description: T("Create an app manifest for an app that has been pushed successfully"),
		Usage: []string{
			T("CF_NAME create-app-manifest APP_NAME [-p /path/to/<app-name>-manifest.yml ]"),
		},
		Flags: fs,
	}
}

func (cmd *CreateAppManifest) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + command_registry.Commands.CommandUsage("create-app-manifest"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs
}

func (cmd *CreateAppManifest) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.appSummaryRepo = deps.RepoLocator.GetAppSummaryRepository()
	cmd.stackRepo = deps.RepoLocator.GetStackRepository()
	cmd.manifest = deps.AppManifest
	return cmd
}

func (cmd *CreateAppManifest) Execute(c flags.FlagContext) {
	application, apiErr := cmd.appSummaryRepo.GetSummary(cmd.appReq.GetApplication().Guid)
	if apiErr != nil {
		cmd.ui.Failed(T("Error getting application summary: ") + apiErr.Error())
	}

	stack, err := cmd.stackRepo.FindByGUID(application.StackGuid)
	if err != nil {
		cmd.ui.Failed(T("Error retrieving stack: ") + err.Error())
	}

	application.Stack = &stack

	cmd.ui.Say(T("Creating an app manifest from current settings of app ") + application.Name + " ...")
	cmd.ui.Say("")

	savePath := "./" + application.Name + "_manifest.yml"

	if c.String("p") != "" {
		savePath = c.String("p")
	}

	f, err := os.Create(savePath)
	if err != nil {
		cmd.ui.Failed(T("Error creating manifest file: ") + err.Error())
	}
	defer f.Close()

	cmd.createManifest(application)
	err = cmd.manifest.Save(f)
	if err != nil {
		cmd.ui.Failed(T("Error creating manifest file: ") + err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("Manifest file created successfully at ") + savePath)
	cmd.ui.Say("")
}

func (cmd *CreateAppManifest) createManifest(app models.Application) {
	cmd.manifest.Memory(app.Name, app.Memory)
	cmd.manifest.Instances(app.Name, app.InstanceCount)
	cmd.manifest.Stack(app.Name, app.Stack.Name)

	if len(app.AppPorts) > 0 {
		cmd.manifest.AppPorts(app.Name, app.AppPorts)
	}

	if app.Command != "" {
		cmd.manifest.StartCommand(app.Name, app.Command)
	}

	if app.BuildpackUrl != "" {
		cmd.manifest.BuildpackUrl(app.Name, app.BuildpackUrl)
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
				cmd.manifest.EnvironmentVars(app.Name, envVarKey, app.EnvironmentVars[envVarKey].(string))
			}
		}
	}

	if len(app.Routes) > 0 {
		for i := 0; i < len(app.Routes); i++ {
			cmd.manifest.Domain(app.Name, app.Routes[i].Host, app.Routes[i].Domain.Name)
		}
	}

	if app.DiskQuota != 0 {
		cmd.manifest.DiskQuota(app.Name, app.DiskQuota)
	}

}

func sortEnvVar(vars map[string]interface{}) []string {
	var varsAry []string
	for k := range vars {
		varsAry = append(varsAry, k)
	}
	sort.Strings(varsAry)

	return varsAry
}
