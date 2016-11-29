package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type UpdateQuotaCommand struct {
	RequiredArgs             flag.Quota               `positional-args:"yes"`
	NumAppInstances          int                      `short:"a" description:"Total number of application instances. -1 represents an unlimited amount."`
	AllowPaidServicePlans    bool                     `long:"allow-paid-service-plans" description:"Can provision instances of paid service plans"`
	DisallowPaidServicePlans bool                     `long:"disallow-paid-service-plans" description:"Cannot provision instances of paid service plans"`
	AppInstanceMemory        flag.MemoryWithUnlimited `short:"i" description:"Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G)"`
	TotalMemory              string                   `short:"m" description:"Total amount of memory (e.g. 1024M, 1G, 10G)"`
	NewName                  string                   `short:"n" description:"New name"`
	NumRoutes                int                      `short:"r" description:"Total number of routes"`
	ReservedRoutePorts       int                      `long:"reserved-route-ports" description:"Maximum number of routes that may be created with reserved ports"`
	NumServiceInstances      int                      `short:"s" description:"Total number of service instances"`
	usage                    interface{}              `usage:"CF_NAME update-quota QUOTA [-m TOTAL_MEMORY] [-i INSTANCE_MEMORY] [-n NEW_NAME] [-r ROUTES] [-s SERVICE_INSTANCES] [-a APP_INSTANCES] [--allow-paid-service-plans | --disallow-paid-service-plans] [--reserved-route-ports RESERVED_ROUTE_PORTS]"`
	relatedCommands          interface{}              `related_commands:"org, quota"`
}

func (_ UpdateQuotaCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ UpdateQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
