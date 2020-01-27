package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CreateOrgQuotaActor

type CreateOrgQuotaActor interface {
	CreateOrganizationQuota(name string, limits v7action.QuotaLimits) (v7action.Warnings, error)
}

type CreateOrgQuotaCommand struct {
	RequiredArgs          flag.OrganizationQuota   `positional-args:"yes"`
	NumAppInstances       flag.IntegerLimit        `short:"a" description:"Total number of application instances. (Default: unlimited)."`
	PaidServicePlans      bool                     `long:"allow-paid-service-plans" description:"Allow provisioning instances of paid service plans. (Default: disallowed)."`
	PerProcessMemory      flag.MemoryWithUnlimited `short:"i" description:"Maximum amount of memory a process can have (e.g. 1024M, 1G, 10G). (Default: unlimited)."`
	TotalMemory           flag.MemoryWithUnlimited `short:"m" description:"Total amount of memory all processes can have (e.g. 1024M, 1G, 10G).  -1 represents an unlimited amount. (Default: 0)."`
	TotalRoutes           flag.IntegerLimit        `short:"r" description:"Total number of routes. -1 represents an unlimited amount. (Default: 0)."`
	TotalReservedPorts    flag.IntegerLimit        `long:"reserved-route-ports" description:"Maximum number of routes that may be created with ports. -1 represents an unlimited amount. (Default: 0)."`
	TotalServiceInstances flag.IntegerLimit        `short:"s" description:"Total number of service instances. -1 represents an unlimited amount. (Default: 0)."`
	usage                 interface{}              `usage:"CF_NAME create-org-quota ORG_QUOTA [-m TOTAL_MEMORY] [-i INSTANCE_MEMORY] [-r ROUTES] [-s SERVICE_INSTANCES] [-a APP_INSTANCES] [--allow-paid-service-plans] [--reserved-route-ports RESERVED_ROUTE_PORTS]"`
	relatedCommands       interface{}              `related_commands:"create-org, org-quotas, set-org-quota"`

	UI          command.UI
	Config      command.Config
	Actor       CreateOrgQuotaActor
	SharedActor command.SharedActor
}

func (cmd *CreateOrgQuotaCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
}

func (cmd CreateOrgQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
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
		TotalMemoryInMB:       types.NullInt(cmd.TotalMemory),
		PerProcessMemoryInMB:  types.NullInt(cmd.PerProcessMemory),
		TotalInstances:        types.NullInt(cmd.NumAppInstances),
		PaidServicesAllowed:   cmd.PaidServicePlans,
		TotalServiceInstances: types.NullInt(cmd.TotalServiceInstances),
		TotalRoutes:           types.NullInt(cmd.TotalRoutes),
		TotalReservedPorts:    types.NullInt(cmd.TotalReservedPorts),
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
