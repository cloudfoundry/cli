package application

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type GetHealthCheck struct {
	ui      terminal.UI
	config  coreconfig.Reader
	appReq  requirements.ApplicationRequirement
	appRepo applications.Repository
}

func init() {
	commandregistry.Register(&GetHealthCheck{})
}

func (cmd *GetHealthCheck) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "get-health-check",
		Description: T("Get the health_check_type value of an app"),
		Usage: []string{
			T("CF_NAME get-health-check APP_NAME"),
		},
	}
}

func (cmd *GetHealthCheck) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument\n\n") + commandregistry.Commands.CommandUsage("get-health-check"))
	}

	cmd.ui.Say(T("Getting health_check_type value for ") + terminal.EntityNameColor(fc.Args()[0]))
	cmd.ui.Say("")

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs
}

func (cmd *GetHealthCheck) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	return cmd
}

func (cmd *GetHealthCheck) Execute(fc flags.FlagContext) error {
	app := cmd.appReq.GetApplication()

	cmd.ui.Say(T("health_check_type is ") + terminal.HeaderColor(app.HealthCheckType))
	return nil
}
