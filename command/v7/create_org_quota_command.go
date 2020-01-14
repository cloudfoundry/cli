package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CreateOrgQuotaActor

type CreateOrgQuotaActor interface {
	CreateOrganizationQuota(orgQuotaName string) (v7action.Warnings, error)
}

type CreateOrgQuotaCommand struct {
	RequiredArgs          flag.OrganizationQuota `positional-args:"yes"`
	NumAppInstances       string                 `short:"a" description:"Total number of application instances. (Default: unlimited)."`
	PaidServicePlans      bool                   `long:"allow-paid-service-plans" description:"Allow provisioning instances of paid service plans. (Default: disallowed)."`
	PerProcessMemory      string                 `short:"i" description:"Maximum amount of memory a process can have (e.g. 1024M, 1G, 10G). (Default: unlimited)."`
	TotalMemory           string                 `short:"m" description:"Total amount of memory all processes can have (e.g. 1024M, 1G, 10G).  -1 represents an unlimited amount. (Default: 0)."`
	TotalRoutes           string                 `short:"r" description:"Total number of routes. -1 represents an unlimited amount. (Default: 0)."`
	TotalReservedPorts    string                 `long:"reserved-route-ports" description:"Maximum number of routes that may be created with ports. -1 represents an unlimited amount. (Default: 0)."`
	TotalServiceInstances string                 `short:"s" description:"Total number of service instances. -1 represents an unlimited amount. (Default: 0)."`
	usage                 interface{}            `usage:"CF_NAME create-org-quota ORG_QUOTA [-m TOTAL_MEMORY] [-i INSTANCE_MEMORY] [-r ROUTES] [-s SERVICE_INSTANCES] [-a APP_INSTANCES] [--allow-paid-service-plans] [--reserved-route-ports RESERVED_ROUTE_PORTS]"`
	relatedCommands       interface{}            `related_commands:"create-org, org-quotas, set-org-quota"`

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

	orgQuotaName := cmd.RequiredArgs.OrganizationQuota

	cmd.UI.DisplayTextWithFlavor("Creating org quota {{.OrganizationQuota}} as {{.User}}...",
		map[string]interface{}{
			"User":              user.Name,
			"OrganizationQuota": orgQuotaName,
		})

	warnings, err := cmd.Actor.CreateOrganizationQuota(orgQuotaName)
	cmd.UI.DisplayWarnings(warnings)

	if _, ok := err.(ccerror.OrgQuotaAlreadyExists); ok {
		cmd.UI.DisplayText(err.Error())
		cmd.UI.DisplayOK()
		return nil
	}

	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
