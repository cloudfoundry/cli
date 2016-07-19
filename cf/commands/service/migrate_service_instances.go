package service

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type MigrateServiceInstances struct {
	ui          terminal.UI
	configRepo  coreconfig.Reader
	serviceRepo api.ServiceRepository
}

func init() {
	commandregistry.Register(&MigrateServiceInstances{})
}

func (cmd *MigrateServiceInstances) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force migration without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "migrate-service-instances",
		Description: T("Migrate service instances from one service plan to another"),
		Usage: []string{
			T("CF_NAME migrate-service-instances v1_SERVICE v1_PROVIDER v1_PLAN v2_SERVICE v2_PLAN\n\n"),
			migrateServiceInstanceWarning(),
		},
		Flags: fs,
	}
}

func (cmd *MigrateServiceInstances) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 5 {
		cmd.ui.Failed(T("Incorrect Usage. Requires v1_SERVICE v1_PROVIDER v1_PLAN v2_SERVICE v2_PLAN as arguments\n\n") + commandregistry.Commands.CommandUsage("migrate-service-instances"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 5)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewMaxAPIVersionRequirement("migrate-service-instances", cf.ServiceAuthTokenMaximumAPIVersion),
	}

	return reqs, nil
}

func (cmd *MigrateServiceInstances) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func migrateServiceInstanceWarning() string {
	return T("WARNING: This operation is internal to Cloud Foundry; service brokers will not be contacted and resources for service instances will not be altered. The primary use case for this operation is to replace a service broker which implements the v1 Service Broker API with a broker which implements the v2 API by remapping service instances from v1 plans to v2 plans.  We recommend making the v1 plan private or shutting down the v1 broker to prevent additional instances from being created. Once service instances have been migrated, the v1 services and plans can be removed from Cloud Foundry.")
}

func (cmd *MigrateServiceInstances) Execute(c flags.FlagContext) error {
	v1 := resources.ServicePlanDescription{
		ServiceLabel:    c.Args()[0],
		ServiceProvider: c.Args()[1],
		ServicePlanName: c.Args()[2],
	}
	v2 := resources.ServicePlanDescription{
		ServiceLabel:    c.Args()[3],
		ServicePlanName: c.Args()[4],
	}
	force := c.Bool("f")

	v1GUID, err := cmd.serviceRepo.FindServicePlanByDescription(v1)
	switch err.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		return errors.New(T("Plan {{.ServicePlanName}} cannot be found",
			map[string]interface{}{
				"ServicePlanName": terminal.EntityNameColor(v1.String()),
			}))
	default:
		return err
	}

	v2GUID, err := cmd.serviceRepo.FindServicePlanByDescription(v2)
	switch err.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		return errors.New(T("Plan {{.ServicePlanName}} cannot be found",
			map[string]interface{}{
				"ServicePlanName": terminal.EntityNameColor(v2.String()),
			}))
	default:
		return err
	}

	count, err := cmd.serviceRepo.GetServiceInstanceCountForServicePlan(v1GUID)
	if err != nil {
		return err
	} else if count == 0 {
		return errors.New(T("Plan {{.ServicePlanName}} has no service instances to migrate", map[string]interface{}{"ServicePlanName": terminal.EntityNameColor(v1.String())}))
	}

	cmd.ui.Warn(migrateServiceInstanceWarning())

	serviceInstancesPhrase := pluralizeServiceInstances(count)

	if !force {
		response := cmd.ui.Confirm(
			T("Really migrate {{.ServiceInstanceDescription}} from plan {{.OldServicePlanName}} to {{.NewServicePlanName}}?>",
				map[string]interface{}{
					"ServiceInstanceDescription": serviceInstancesPhrase,
					"OldServicePlanName":         terminal.EntityNameColor(v1.String()),
					"NewServicePlanName":         terminal.EntityNameColor(v2.String()),
				}))
		if !response {
			return nil
		}
	}

	cmd.ui.Say(T("Attempting to migrate {{.ServiceInstanceDescription}}...", map[string]interface{}{"ServiceInstanceDescription": serviceInstancesPhrase}))

	changedCount, err := cmd.serviceRepo.MigrateServicePlanFromV1ToV2(v1GUID, v2GUID)
	if err != nil {
		return err
	}

	cmd.ui.Say(T("{{.CountOfServices}} migrated.", map[string]interface{}{"CountOfServices": pluralizeServiceInstances(changedCount)}))
	cmd.ui.Ok()

	return nil
}

func pluralizeServiceInstances(count int) string {
	var phrase string
	if count == 1 {
		phrase = T("service instance")
	} else {
		phrase = T("service instances")
	}

	return fmt.Sprintf("%d %s", count, phrase)
}
