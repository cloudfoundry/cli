package application

import (
	"errors"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type RestartAppInstance struct {
	ui               terminal.UI
	config           coreconfig.Reader
	appReq           requirements.ApplicationRequirement
	appInstancesRepo appinstances.Repository
}

func init() {
	commandregistry.Register(&RestartAppInstance{})
}

func (cmd *RestartAppInstance) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "restart-app-instance",
		Description: T("Terminate the running application Instance at the given index and instantiate a new instance of the application with the same index"),
		Usage: []string{
			T("CF_NAME restart-app-instance APP_NAME INDEX"),
		},
	}
}

func (cmd *RestartAppInstance) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		usage := commandregistry.Commands.CommandUsage("restart-app-instance")
		cmd.ui.Failed(T("Incorrect Usage. Requires arguments\n\n") + usage)
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	appName := fc.Args()[0]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(appName)

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *RestartAppInstance) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appInstancesRepo = deps.RepoLocator.GetAppInstancesRepository()
	return cmd
}

func (cmd *RestartAppInstance) Execute(fc flags.FlagContext) error {
	app := cmd.appReq.GetApplication()

	instance, err := strconv.Atoi(fc.Args()[1])

	if err != nil {
		return errors.New(T("Instance must be a non-negative integer"))
	}

	cmd.ui.Say(T("Restarting instance {{.Instance}} of application {{.AppName}} as {{.Username}}",
		map[string]interface{}{
			"Instance": instance,
			"AppName":  terminal.EntityNameColor(app.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.appInstancesRepo.DeleteInstance(app.GUID, instance)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
	return nil
}
