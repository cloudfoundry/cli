package application

import (
	"encoding/json"
	"sort"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Env struct {
	ui      terminal.UI
	config  core_config.Reader
	appRepo applications.ApplicationRepository
}

func NewEnv(ui terminal.UI, config core_config.Reader, appRepo applications.ApplicationRepository) (cmd *Env) {
	cmd = new(Env)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo
	return
}

func (cmd *Env) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "env",
		ShortName:   "e",
		Description: T("Show all env variables for an app"),
		Usage:       T("CF_NAME env APP"),
	}
}

func (cmd *Env) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) ([]requirements.Requirement, error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *Env) Run(c *cli.Context) {
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

	env, err := cmd.appRepo.ReadEnv(app.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	cmd.displaySystemProvidedEnvironment(env.System)
	cmd.ui.Say("")
	cmd.displayUserProvidedEnvironment(env.Environment)
	cmd.ui.Say("")
	cmd.displayRunningEnvironment(env.Running)
	cmd.ui.Say("")
	cmd.displayStagingEnvironment(env.Staging)
	cmd.ui.Say("")
}

func (cmd *Env) displaySystemProvidedEnvironment(env map[string]interface{}) {
	var vcapServices string
	servicesAsMap, ok := env["VCAP_SERVICES"].(map[string]interface{})
	if ok && len(servicesAsMap) > 0 {
		jsonBytes, err := json.MarshalIndent(env, "", " ")
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		vcapServices = string(jsonBytes)
	}

	if len(vcapServices) == 0 {
		cmd.ui.Say(T("No system-provided env variables have been set"))
		return
	}
	cmd.ui.Say(terminal.EntityNameColor(T("System-Provided:")))
	cmd.ui.Say(vcapServices)
}

func (cmd *Env) displayUserProvidedEnvironment(envVars map[string]interface{}) {
	if len(envVars) == 0 {
		cmd.ui.Say(T("No user-defined env variables have been set"))
		return
	}

	keys := make([]string, 0, len(envVars))
	for key, _ := range envVars {
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
	for key, _ := range envVars {
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
	for key, _ := range envVars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	cmd.ui.Say(terminal.EntityNameColor(T("Staging Environment Variable Groups:")))
	for _, key := range keys {
		cmd.ui.Say("%s: %v", key, envVars[key])
	}
}
