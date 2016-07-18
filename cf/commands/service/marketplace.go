package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/actors/servicebuilder"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type MarketplaceServices struct {
	ui             terminal.UI
	config         coreconfig.Reader
	serviceBuilder servicebuilder.ServiceBuilder
}

func init() {
	commandregistry.Register(&MarketplaceServices{})
}

func (cmd *MarketplaceServices) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["s"] = &flags.StringFlag{ShortName: "s", Usage: T("Show plan details for a particular service offering")}

	return commandregistry.CommandMetadata{
		Name:        "marketplace",
		ShortName:   "m",
		Description: T("List available offerings in the marketplace"),
		Usage: []string{
			"CF_NAME marketplace ",
			fmt.Sprintf("[-s %s] ", T("SERVICE")),
		},
		Flags: fs,
	}
}

func (cmd *MarketplaceServices) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewAPIEndpointRequirement(),
	}

	return reqs
}

func (cmd *MarketplaceServices) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceBuilder = deps.ServiceBuilder
	return cmd
}

func (cmd *MarketplaceServices) Execute(c flags.FlagContext) error {
	serviceName := c.String("s")

	var err error
	if serviceName != "" {
		err = cmd.marketplaceByService(serviceName)
	} else {
		err = cmd.marketplace()
	}
	if err != nil {
		return err
	}

	return nil
}

func (cmd MarketplaceServices) marketplaceByService(serviceName string) error {
	var serviceOffering models.ServiceOffering
	var err error

	if cmd.config.HasSpace() {
		cmd.ui.Say(T("Getting service plan information for service {{.ServiceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"ServiceName": terminal.EntityNameColor(serviceName),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
			}))
		serviceOffering, err = cmd.serviceBuilder.GetServiceByNameForSpaceWithPlans(serviceName, cmd.config.SpaceFields().GUID)
	} else if !cmd.config.IsLoggedIn() {
		cmd.ui.Say(T("Getting service plan information for service {{.ServiceName}}...", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceName)}))
		serviceOffering, err = cmd.serviceBuilder.GetServiceByNameWithPlans(serviceName)
	} else {
		err = errors.New(T("Cannot list plan information for {{.ServiceName}} without a targeted space",
			map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceName)}))
	}
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if serviceOffering.GUID == "" {
		cmd.ui.Say(T("Service offering not found"))
		return nil
	}

	table := cmd.ui.Table([]string{T("service plan"), T("description"), T("free or paid")})
	for _, plan := range serviceOffering.Plans {
		var freeOrPaid string
		if plan.Free {
			freeOrPaid = "free"
		} else {
			freeOrPaid = "paid"
		}
		table.Add(plan.Name, plan.Description, freeOrPaid)
	}

	table.Print()
	return nil
}

func (cmd MarketplaceServices) marketplace() error {
	var serviceOfferings models.ServiceOfferings
	var err error

	if cmd.config.HasSpace() {
		cmd.ui.Say(T("Getting services from marketplace in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
			}))
		serviceOfferings, err = cmd.serviceBuilder.GetServicesForSpaceWithPlans(cmd.config.SpaceFields().GUID)
	} else if !cmd.config.IsLoggedIn() {
		cmd.ui.Say(T("Getting all services from marketplace..."))
		serviceOfferings, err = cmd.serviceBuilder.GetAllServicesWithPlans()
	} else {
		err = errors.New(T("Cannot list marketplace services without a targeted space"))
	}
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(serviceOfferings) == 0 {
		cmd.ui.Say(T("No service offerings found"))
		return nil
	}

	table := cmd.ui.Table([]string{T("service"), T("plans"), T("description")})

	sort.Sort(serviceOfferings)
	var paidPlanExists bool
	for _, offering := range serviceOfferings {
		planNames := ""

		for _, plan := range offering.Plans {
			if plan.Name == "" {
				continue
			}
			if plan.Free {
				planNames += ", " + plan.Name
			} else {
				paidPlanExists = true
				planNames += ", " + plan.Name + "*"
			}
		}

		planNames = strings.TrimPrefix(planNames, ", ")

		table.Add(offering.Label, planNames, offering.Description)
	}

	table.Print()
	if paidPlanExists {
		cmd.ui.Say(T("\n* These service plans have an associated cost. Creating a service instance will incur this cost."))
	}
	cmd.ui.Say(T("\nTIP:  Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
	return nil
}
