package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/types"
)

type CreateOrgQuotaCommand struct {
	BaseCommand

	RequiredArgs          flag.OrganizationQuota      `positional-args:"yes"`
	NumAppInstances       flag.IntegerLimit           `short:"a" description:"Total number of application instances. (Default: unlimited)."`
	PaidServicePlans      bool                        `long:"allow-paid-service-plans" description:"Allow provisioning instances of paid service plans. (Default: disallowed)."`
	PerProcessMemory      flag.MegabytesWithUnlimited `short:"i" description:"Maximum amount of memory a process can have (e.g. 1024M, 1G, 10G). (Default: unlimited)."`
	TotalMemory           flag.MegabytesWithUnlimited `short:"m" description:"Total amount of memory all processes can have (e.g. 1024M, 1G, 10G).  -1 represents an unlimited amount. (Default: 0)."`
	TotalRoutes           flag.IntegerLimit           `short:"r" description:"Total number of routes. -1 represents an unlimited amount. (Default: 0)."`
	TotalReservedPorts    flag.IntegerLimit           `long:"reserved-route-ports" description:"Maximum number of routes that may be created with ports. -1 represents an unlimited amount. (Default: 0)."`
	TotalServiceInstances flag.IntegerLimit           `short:"s" description:"Total number of service instances. -1 represents an unlimited amount. (Default: 0)."`
	TotalLogVolume        flag.BytesWithUnlimited     `short:"l" description:"Total log volume per second all processes can have, in bytes (e.g. 128B, 4K, 1M). -1 represents an unlimited amount. (Default: -1)."`

	usage           interface{} `usage:"CF_NAME create-org-quota ORG_QUOTA [-m TOTAL_MEMORY] [-i INSTANCE_MEMORY] [-r ROUTES] [-s SERVICE_INSTANCES] [-a APP_INSTANCES] [--allow-paid-service-plans] [--reserved-route-ports RESERVED_ROUTE_PORTS] [-l LOG_VOLUME]"`
	relatedCommands interface{} `related_commands:"create-org, org-quotas, set-org-quota"`
}

func (cmd CreateOrgQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	orgQuotaName := cmd.RequiredArgs.OrganizationQuotaName

	cmd.UI.DisplayTextWithFlavor("Creating org quota {{.OrganizationQuotaName}} as {{.User}}...",
		map[string]interface{}{
			"User":                  user.Name,
			"OrganizationQuotaName": orgQuotaName,
		})

	warnings, err := cmd.Actor.CreateOrganizationQuota(orgQuotaName, v7action.QuotaLimits{
		TotalMemoryInMB:       convertMegabytesFlagToNullInt(cmd.TotalMemory),
		PerProcessMemoryInMB:  convertMegabytesFlagToNullInt(cmd.PerProcessMemory),
		TotalInstances:        convertIntegerLimitFlagToNullInt(cmd.NumAppInstances),
		PaidServicesAllowed:   &cmd.PaidServicePlans,
		TotalServiceInstances: convertIntegerLimitFlagToNullInt(cmd.TotalServiceInstances),
		TotalRoutes:           convertIntegerLimitFlagToNullInt(cmd.TotalRoutes),
		TotalReservedPorts:    convertIntegerLimitFlagToNullInt(cmd.TotalReservedPorts),
		TotalLogVolume:        convertBytesFlagToNullInt(cmd.TotalLogVolume),
	})
	cmd.UI.DisplayWarnings(warnings)

	if _, ok := err.(ccerror.QuotaAlreadyExists); ok {
		cmd.UI.DisplayWarning(err.Error())
		cmd.UI.DisplayOK()
		return nil
	}

	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}

func convertBytesFlagToNullInt(flag flag.BytesWithUnlimited) *types.NullInt {
	if !flag.IsSet {
		return nil
	}

	return &types.NullInt{IsSet: true, Value: flag.Value}
}

func convertMegabytesFlagToNullInt(flag flag.MegabytesWithUnlimited) *types.NullInt {
	if !flag.IsSet {
		return nil
	}

	return &types.NullInt{IsSet: true, Value: flag.Value}
}

func convertIntegerLimitFlagToNullInt(flag flag.IntegerLimit) *types.NullInt {
	if !flag.IsSet {
		return nil
	}

	return &types.NullInt{IsSet: true, Value: flag.Value}
}
