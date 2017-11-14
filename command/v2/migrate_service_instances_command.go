package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type MigrateServiceInstancesCommand struct {
	RequiredArgs flag.MigrateServiceInstancesArgs `positional-args:"yes"`
	Force        bool                             `short:"f" description:"Force migration without confirmation"`
	usage        interface{}                      `usage:"CF_NAME migrate-service-instances v1_SERVICE v1_PROVIDER v1_PLAN v2_SERVICE v2_PLAN\n\nWARNING: This operation is internal to Cloud Foundry; service brokers will not be contacted and resources for service instances will not be altered. The primary use case for this operation is to replace a service broker which implements the v1 Service Broker API with a broker which implements the v2 API by remapping service instances from v1 plans to v2 plans.  We recommend making the v1 plan private or shutting down the v1 broker to prevent additional instances from being created. Once service instances have been migrated, the v1 services and plans can be removed from Cloud Foundry."`
}

func (MigrateServiceInstancesCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (MigrateServiceInstancesCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
