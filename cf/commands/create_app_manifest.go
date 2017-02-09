package commands

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/api/stacks"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/manifest"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type CreateAppManifest struct {
	ui               terminal.UI
	config           coreconfig.Reader
	appSummaryRepo   api.AppSummaryRepository
	stackRepo        stacks.StackRepository
	appInstancesRepo appinstances.Repository
	appReq           requirements.ApplicationRequirement
	manifest         manifest.App
}

func init() {
	commandregistry.Register(&CreateAppManifest{})
}

func (cmd *CreateAppManifest) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Specify a path for file creation. If path not specified, manifest file is created in current working directory.")}

	return commandregistry.CommandMetadata{
		Name:        "create-app-manifest",
		Description: T("Create an app manifest for an app that has been pushed successfully"),
		Usage: []string{
			T("CF_NAME create-app-manifest APP_NAME [-p /path/to/<app-name>-manifest.yml ]"),
		},
		Flags: fs,
	}
}

func (cmd *CreateAppManifest) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + commandregistry.Commands.CommandUsage("create-app-manifest"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *CreateAppManifest) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appSummaryRepo = deps.RepoLocator.GetAppSummaryRepository()
	cmd.stackRepo = deps.RepoLocator.GetStackRepository()
	cmd.manifest = deps.AppManifest
	return cmd
}

func (cmd *CreateAppManifest) Execute(c flags.FlagContext) error {
	application, apiErr := cmd.appSummaryRepo.GetSummary(cmd.appReq.GetApplication().GUID)
	if apiErr != nil {
		return errors.New(T("Error getting application summary: ") + apiErr.Error())
	}

	stack, err := cmd.stackRepo.FindByGUID(application.StackGUID)
	if err != nil {
		return errors.New(T("Error retrieving stack: ") + err.Error())
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
		return errors.New(T("Error creating manifest file: ") + err.Error())
	}
	defer f.Close()

	err = cmd.createManifest(application)
	if err != nil {
		return err
	}
	err = cmd.manifest.Save(f)
	if err != nil {
		return errors.New(T("Error creating manifest file: ") + err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("Manifest file created successfully at ") + savePath)
	cmd.ui.Say("")
	return nil
}

func (cmd *CreateAppManifest) createManifest(app models.Application) error {
	cmd.manifest.Memory(app.Name, app.Memory)
	cmd.manifest.Instances(app.Name, app.InstanceCount)
	cmd.manifest.Stack(app.Name, app.Stack.Name)

	if len(app.AppPorts) > 0 {
		cmd.manifest.AppPorts(app.Name, app.AppPorts)
	}

	if app.Command != "" {
		cmd.manifest.StartCommand(app.Name, app.Command)
	}

	if app.BuildpackURL != "" {
		cmd.manifest.BuildpackURL(app.Name, app.BuildpackURL)
	}

	if len(app.Services) > 0 {
		for _, service := range app.Services {
			cmd.manifest.Service(app.Name, service.Name)
		}
	}

	if app.HealthCheckTimeout > 0 {
		cmd.manifest.HealthCheckTimeout(app.Name, app.HealthCheckTimeout)
	}

	if app.HealthCheckType != "port" {
		cmd.manifest.HealthCheckType(app.Name, app.HealthCheckType)
	}

	if app.HealthCheckType == "http" &&
		app.HealthCheckHTTPEndpoint != "" &&
		app.HealthCheckHTTPEndpoint != "/" {
		cmd.manifest.HealthCheckHTTPEndpoint(app.Name, app.HealthCheckHTTPEndpoint)
	}

	if len(app.EnvironmentVars) > 0 {
		sorted := sortEnvVar(app.EnvironmentVars)
		for _, envVarKey := range sorted {
			switch app.EnvironmentVars[envVarKey].(type) {
			default:
				return errors.New(T("Failed to create manifest, unable to parse environment variable: ") + envVarKey)
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
			cmd.manifest.Route(app.Name, app.Routes[i].Host, app.Routes[i].Domain.Name, app.Routes[i].Path, app.Routes[i].Port)
		}
	}

	if app.DiskQuota != 0 {
		cmd.manifest.DiskQuota(app.Name, app.DiskQuota)
	}

	return nil
}

func sortEnvVar(vars map[string]interface{}) []string {
	var varsAry []string
	for k := range vars {
		varsAry = append(varsAry, k)
	}
	sort.Strings(varsAry)

	return varsAry
}
