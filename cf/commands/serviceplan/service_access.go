package serviceplan

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ServiceAccess struct {
	ui     terminal.UI
	config configuration.Reader
	actor  actors.ServiceActor
}

func NewServiceAccess(ui terminal.UI, config configuration.Reader, actor actors.ServiceActor) (cmd *ServiceAccess) {
	return &ServiceAccess{
		ui:     ui,
		config: config,
		actor:  actor,
	}
}

func (cmd *ServiceAccess) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "service-access",
		Description: T("List service access settings"),
		Usage:       "CF_NAME service-access [-b BROKER] [-e SERVICE] [-o ORG]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("b", T("access for plans of a particular broker")),
			flag_helpers.NewStringFlag("e", T("access for plans of a particular service offering")),
			flag_helpers.NewStringFlag("o", T("plans accessible by a particular organization")),
		},
	}
}

func (cmd *ServiceAccess) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *ServiceAccess) Run(c *cli.Context) {
	brokers, err := cmd.actor.FilterBrokers(c.String("b"), c.String("e"), c.String("o"))
	if err != nil {
		cmd.ui.Failed(T("Failed fetching service brokers.\n%s"), err)
		return
	}
	cmd.printTable(brokers)
}

func (cmd ServiceAccess) printTable(brokers []models.ServiceBroker) {
	for _, serviceBroker := range brokers {
		cmd.ui.Say(fmt.Sprintf(T("broker: {{.Name}}", map[string]interface{}{"Name": serviceBroker.Name})))

		table := terminal.NewTable(cmd.ui, []string{"", T("service"), T("plan"), T("access"), T("orgs")})
		for _, service := range serviceBroker.Services {
			if len(service.Plans) > 0 {
				for _, plan := range service.Plans {
					table.Add("", service.Label, plan.Name, cmd.formatAccess(plan.Public, plan.OrgNames), strings.Join(plan.OrgNames, ","))
				}
			} else {
				table.Add("", service.Label, "", "", "")
			}
		}
		table.Print()

		cmd.ui.Say("")
	}
	return
}

func (cmd ServiceAccess) formatAccess(public bool, orgNames []string) string {
	if public {
		return T("all")
	}
	if len(orgNames) > 0 {
		return T("limited")
	}
	return T("none")
}
