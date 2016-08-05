package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type UpdateSpaceQuotaCommand struct {
	RequiredArgs             flags.SpaceQuota `positional-args:"yes"`
	AppInstanceMemory        string           `short:"i" description:"Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount. (Default: unlimited)"`
	TotalMemory              string           `short:"m" description:"Total amount of memory a space can have (e.g. 1024M, 1G, 10G)"`
	Name                     string           `short:"n" description:"New name"`
	NumRoutes                int              `short:"r" description:"Total number of routes"`
	NumServiceInstances      int              `short:"s" description:"Total number of service instances"`
	AllowPaidServicePlans    bool             `long:"allow-paid-service-plans" description:"Can provision instances of paid service plans"`
	DisallowPaidServicePlans bool             `long:"disallow-paid-service-plans" description:"Can not provision instances of paid service plans"`
	NumAppInstances          int              `short:"a" description:"Total number of application instances. -1 represents an unlimited amount. (Default: unlimited)"`
	ReservedRoutePorts       int              `long:"reserved-route-ports" description:"Maximum number of routes that may be created with reserved ports (Default: 0)"`
}

func (_ UpdateSpaceQuotaCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
