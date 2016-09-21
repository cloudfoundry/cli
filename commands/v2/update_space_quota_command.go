package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UpdateSpaceQuotaCommand struct {
	RequiredArgs             flags.SpaceQuota          `positional-args:"yes"`
	NumAppInstances          int                       `short:"a" description:"Total number of application instances. -1 represents an unlimited amount."`
	AllowPaidServicePlans    bool                      `long:"allow-paid-service-plans" description:"Can provision instances of paid service plans"`
	DisallowPaidServicePlans bool                      `long:"disallow-paid-service-plans" description:"Can not provision instances of paid service plans"`
	AppInstanceMemory        flags.MemoryWithUnlimited `short:"i" description:"Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount."`
	TotalMemory              string                    `short:"m" description:"Total amount of memory a space can have (e.g. 1024M, 1G, 10G)"`
	Name                     string                    `short:"n" description:"New name"`
	NumRoutes                int                       `short:"r" description:"Total number of routes"`
	ReservedRoutePorts       int                       `long:"reserved-route-ports" description:"Maximum number of routes that may be created with reserved ports"`
	NumServiceInstances      int                       `short:"s" description:"Total number of service instances"`
	usage                    interface{}               `usage:"CF_NAME update-space-quota SPACE_QUOTA [-i INSTANCE_MEMORY] [-m MEMORY] [-n NAME] [-r ROUTES] [-s SERVICE_INSTANCES] [-a APP_INSTANCES] [--allow-paid-service-plans | --disallow-paid-service-plans] [--reserved-route-ports RESERVED_ROUTE_PORTS]"`
	relatedCommands          interface{}               `related_commands:"space-quota, space-quotas"`
}

func (_ UpdateSpaceQuotaCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ UpdateSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
