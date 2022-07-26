package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type UpdateSpaceQuotaCommand struct {
	BaseCommand

	RequiredArgs          flag.SpaceQuota             `positional-args:"Yes"`
	NumAppInstances       flag.IntegerLimit           `short:"a" description:"Total number of application instances. -1 represents an unlimited amount."`
	PaidServicePlans      bool                        `long:"allow-paid-service-plans" description:"Allow provisioning instances of paid service plans."`
	NoPaidServicePlans    bool                        `long:"disallow-paid-service-plans" description:"Disallow provisioning instances of paid service plans."`
	PerProcessMemory      flag.MegabytesWithUnlimited `short:"i" description:"Maximum amount of memory a process can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount."`
	TotalMemory           flag.MegabytesWithUnlimited `short:"m" description:"Total amount of memory all processes can have (e.g. 1024M, 1G, 10G).  -1 represents an unlimited amount."`
	NewName               string                      `short:"n" description:"New name"`
	TotalRoutes           flag.IntegerLimit           `short:"r" description:"Total number of routes. -1 represents an unlimited amount."`
	TotalReservedPorts    flag.IntegerLimit           `long:"reserved-route-ports" description:"Maximum number of routes that may be created with ports. -1 represents an unlimited amount."`
	TotalServiceInstances flag.IntegerLimit           `short:"s" description:"Total number of service instances. -1 represents an unlimited amount."`
	TotalLogVolume        flag.BytesWithUnlimited     `short:"l" description:"Total log volume per second all processes can have, in bytes (e.g. 128B, 4K, 1M). -1 represents an unlimited amount."`

	usage           interface{} `usage:"CF_NAME update-space-quota QUOTA [-m TOTAL_MEMORY] [-i INSTANCE_MEMORY] [-n NEW_NAME] [-r ROUTES] [-s SERVICE_INSTANCES] [-a APP_INSTANCES] [--allow-paid-service-plans | --disallow-paid-service-plans] [--reserved-route-ports RESERVED_ROUTE_PORTS] [-l LOG_VOLUME]"`
	relatedCommands interface{} `related_commands:"space, space-quota, space-quotas"`
}

func (cmd UpdateSpaceQuotaCommand) Execute(args []string) error {
	if cmd.PaidServicePlans && cmd.NoPaidServicePlans {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--allow-paid-service-plans", "--disallow-paid-service-plans"},
		}
	}

	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	oldQuotaName := cmd.RequiredArgs.SpaceQuota
	orgGUID := cmd.Config.TargetedOrganization().GUID

	cmd.UI.DisplayTextWithFlavor("Updating space quota {{.QuotaName}} for org {{.OrgName}} as {{.User}}...",
		map[string]interface{}{
			"QuotaName": oldQuotaName,
			"OrgName":   cmd.Config.TargetedOrganizationName(),
			"User":      user.Name,
		})

	var paidServicesAllowed *bool
	if cmd.PaidServicePlans || cmd.NoPaidServicePlans {
		paidServicesAllowed = &cmd.PaidServicePlans
	}

	updatedQuotaLimits := v7action.QuotaLimits{
		TotalMemoryInMB:       convertMegabytesFlagToNullInt(cmd.TotalMemory),
		PerProcessMemoryInMB:  convertMegabytesFlagToNullInt(cmd.PerProcessMemory),
		TotalInstances:        convertIntegerLimitFlagToNullInt(cmd.NumAppInstances),
		PaidServicesAllowed:   paidServicesAllowed,
		TotalServiceInstances: convertIntegerLimitFlagToNullInt(cmd.TotalServiceInstances),
		TotalRoutes:           convertIntegerLimitFlagToNullInt(cmd.TotalRoutes),
		TotalReservedPorts:    convertIntegerLimitFlagToNullInt(cmd.TotalReservedPorts),
		TotalLogVolume:        convertBytesFlagToNullInt(cmd.TotalLogVolume),
	}

	warnings, err := cmd.Actor.UpdateSpaceQuota(oldQuotaName, orgGUID, cmd.NewName, updatedQuotaLimits)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
