package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CreateSpaceQuotaCommand struct {
	RequiredArgs                flag.SpaceQuota          `positional-args:"yes"`
	NumAppInstances             int                      `short:"a" description:"Total number of application instances. -1 represents an unlimited amount. (Default: unlimited)"`
	AllowPaidServicePlans       bool                     `long:"allow-paid-service-plans" description:"Can provision instances of paid service plans (Default: disallowed)"`
	IndividualAppInstanceMemory flag.MemoryWithUnlimited `short:"i" description:"Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount. (Default: unlimited)"`
	TotalMemory                 string                   `short:"m" description:"Total amount of memory a space can have (e.g. 1024M, 1G, 10G)"`
	NumRoutes                   int                      `short:"r" description:"Total number of routes"`
	ReservedRoutePorts          int                      `long:"reserved-route-ports" description:"Maximum number of routes that may be created with reserved ports (Default: 0)"`
	NumServiceInstances         int                      `short:"s" description:"Total number of service instances"`
	usage                       interface{}              `usage:"CF_NAME create-space-quota QUOTA [-i INSTANCE_MEMORY] [-m MEMORY] [-r ROUTES] [-s SERVICE_INSTANCES] [-a APP_INSTANCES] [--allow-paid-service-plans] [--reserved-route-ports RESERVED_ROUTE_PORTS]"`
	relatedCommands             interface{}              `related_commands:"quotas, space-quotas"`
}

func (CreateSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (CreateSpaceQuotaCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
