package service

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DeleteService struct {
	ui                 terminal.UI
	config             coreconfig.Reader
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&DeleteService{})
}

func (cmd *DeleteService) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-service",
		ShortName:   "ds",
		Description: T("Delete a service instance"),
		Usage: []string{
			T("CF_NAME delete-service SERVICE_INSTANCE [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-service"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *DeleteService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func (cmd *DeleteService) Execute(c flags.FlagContext) error {
	serviceName := c.Args()[0]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("service"), serviceName) {
			return nil
		}
	}

	cmd.ui.Say(T("Deleting service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceName),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	instance, err := cmd.serviceRepo.FindInstanceByName(serviceName)

	switch err.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Service {{.ServiceName}} does not exist.", map[string]interface{}{"ServiceName": serviceName}))
		return nil
	default:
		return err
	}

	err = cmd.serviceRepo.DeleteService(instance)
	if err != nil {
		return err
	}

	err = printSuccessMessageForServiceInstance(serviceName, cmd.serviceRepo, cmd.ui)
	if err != nil {
		cmd.ui.Ok()
	}
	return nil
}
