package application

import (
	"encoding/json"
	"sort"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type Env struct {
	ui      terminal.UI
	config  coreconfig.Reader
	appRepo applications.ApplicationRepository
}

func init() {
	commandregistry.Register(&Env{})
}

func (cmd *Env) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "env",
		ShortName:   "e",
		Description: T("Show all env variables for an app"),
		Usage: []string{
			T("CF_NAME env APP_NAME"),
		},
	}
}

func (cmd *Env) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("env"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return reqs
}

func (cmd *Env) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *Env) Execute(c flags.FlagContext) {
	app, err := cmd.appRepo.Read(c.Args()[0])
	if notFound, ok := err.(*errors.ModelNotFoundError); ok {
		cmd.ui.Failed(notFound.Error())
	}

	cmd.ui.Say(T("Getting env variables for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	env, err := cmd.appRepo.ReadEnv(app.GUID)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	cmd.displaySystemiAndAppProvidedEnvironment(env.System, env.Application)
	cmd.ui.Say("")
	cmd.displayUserProvidedEnvironment(env.Environment)
	cmd.ui.Say("")
	cmd.displayRunningEnvironment(env.Running)
	cmd.ui.Say("")
	cmd.displayStagingEnvironment(env.Staging)
	cmd.ui.Say("")
}

func (cmd *Env) displaySystemiAndAppProvidedEnvironment(env map[string]interface{}, app map[string]interface{}) {
	var vcapServices string
	var vcapApplication string

	servicesAsMap, ok := env["VCAP_SERVICES"].(map[string]interface{})
	if ok && len(servicesAsMap) > 0 {
		jsonBytes, err := json.MarshalIndent(env, "", " ")
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		vcapServices = string(jsonBytes)
	}

	applicationAsMap, ok := app["VCAP_APPLICATION"].(map[string]interface{})
	if ok && len(applicationAsMap) > 0 {
		jsonBytes, err := json.MarshalIndent(app, "", " ")
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		vcapApplication = string(jsonBytes)
	}

	if len(vcapServices) == 0 && len(vcapApplication) == 0 {
		cmd.ui.Say(T("No system-provided env variables have been set"))
		return
	}

	cmd.ui.Say(terminal.EntityNameColor(T("System-Provided:")))

	cmd.ui.Say(vcapServices)
	cmd.ui.Say("")
	cmd.ui.Say(vcapApplication)
}

func (cmd *Env) displayUserProvidedEnvironment(envVars map[string]interface{}) {
	if len(envVars) == 0 {
		cmd.ui.Say(T("No user-defined env variables have been set"))
		return
	}

	keys := make([]string, 0, len(envVars))
	for key := range envVars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	cmd.ui.Say(terminal.EntityNameColor(T("User-Provided:")))
	for _, key := range keys {
		cmd.ui.Say("%s: %v", key, envVars[key])
	}
}

func (cmd *Env) displayRunningEnvironment(envVars map[string]interface{}) {
	if len(envVars) == 0 {
		cmd.ui.Say(T("No running env variables have been set"))
		return
	}

	keys := make([]string, 0, len(envVars))
	for key := range envVars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	cmd.ui.Say(terminal.EntityNameColor(T("Running Environment Variable Groups:")))
	for _, key := range keys {
		cmd.ui.Say("%s: %v", key, envVars[key])
	}
}

func (cmd *Env) displayStagingEnvironment(envVars map[string]interface{}) {
	if len(envVars) == 0 {
		cmd.ui.Say(T("No staging env variables have been set"))
		return
	}

	keys := make([]string, 0, len(envVars))
	for key := range envVars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	cmd.ui.Say(terminal.EntityNameColor(T("Staging Environment Variable Groups:")))
	for _, key := range keys {
		cmd.ui.Say("%s: %v", key, envVars[key])
	}
}
