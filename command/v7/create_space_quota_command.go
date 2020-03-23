package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateSpaceQuotaCommand struct {
	BaseCommand

	RequiredArgs          flag.SpaceQuota          `positional-args:"yes"`
	NumAppInstances       flag.IntegerLimit        `short:"a" description:"Total number of application instances. (Default: unlimited)."`
	PaidServicePlans      bool                     `long:"allow-paid-service-plans" description:"Allow provisioning instances of paid service plans. (Default: disallowed)."`
	PerProcessMemory      flag.MemoryWithUnlimited `short:"i" description:"Maximum amount of memory a process can have (e.g. 1024M, 1G, 10G). (Default: unlimited)."`
	TotalMemory           flag.MemoryWithUnlimited `short:"m" description:"Total amount of memory all processes can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount. (Default: 0)."`
	TotalRoutes           flag.IntegerLimit        `short:"r" description:"Total number of routes. -1 represents an unlimited amount. (Default: 0)."`
	TotalReservedPorts    flag.IntegerLimit        `long:"reserved-route-ports" description:"Maximum number of routes that may be created with ports. -1 represents an unlimited amount. (Default: 0)."`
	TotalServiceInstances flag.IntegerLimit        `short:"s" description:"Total number of service instances. -1 represents an unlimited amount. (Default: 0)."`
	usage                 interface{}              `usage:"CF_NAME create-space-quota QUOTA [-m TOTAL_MEMORY] [-i INSTANCE_MEMORY] [-r ROUTES] [-s SERVICE_INSTANCES] [-a APP_INSTANCES] [--allow-paid-service-plans] [--reserved-route-ports RESERVED_ROUTE_PORTS]"`
	relatedCommands       interface{}              `related_commands:"create-space, space-quotas, set-space-quota"`
}

func (cmd CreateSpaceQuotaCommand) Execute([]string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating space quota {{.SpaceQuota}} for org {{.CurrentOrg}} as {{.CurrentUser}}...", map[string]interface{}{
		"SpaceQuota":  cmd.RequiredArgs.SpaceQuota,
		"CurrentOrg":  cmd.Config.TargetedOrganization().Name,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.CreateSpaceQuota(
		cmd.RequiredArgs.SpaceQuota,
		cmd.Config.TargetedOrganization().GUID,
		v7action.QuotaLimits{
			TotalMemoryInMB:       convertMegabytesFlagToNullInt(cmd.TotalMemory),
			PerProcessMemoryInMB:  convertMegabytesFlagToNullInt(cmd.PerProcessMemory),
			TotalInstances:        convertIntegerLimitFlagToNullInt(cmd.NumAppInstances),
			PaidServicesAllowed:   &cmd.PaidServicePlans,
			TotalServiceInstances: convertIntegerLimitFlagToNullInt(cmd.TotalServiceInstances),
			TotalRoutes:           convertIntegerLimitFlagToNullInt(cmd.TotalRoutes),
			TotalReservedPorts:    convertIntegerLimitFlagToNullInt(cmd.TotalReservedPorts),
		},
	)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case ccerror.QuotaAlreadyExists:
			cmd.UI.DisplayWarning(err.Error())
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
