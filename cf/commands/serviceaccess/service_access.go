package serviceaccess

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ServiceAccess struct {
	ui             terminal.UI
	config         core_config.Reader
	actor          actors.ServiceActor
	tokenRefresher authentication.TokenRefresher
}

func NewServiceAccess(ui terminal.UI, config core_config.Reader, actor actors.ServiceActor, tokenRefresher authentication.TokenRefresher) (cmd *ServiceAccess) {
	return &ServiceAccess{
		ui:             ui,
		config:         config,
		actor:          actor,
		tokenRefresher: tokenRefresher,
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
		requirementsFactory.NewMinCCApiVersionRequirement(cmd.Metadata().Name, 2, 13, 0),
	}
	return
}

func (cmd *ServiceAccess) Run(c *cli.Context) {
	_, err := cmd.tokenRefresher.RefreshAuthToken()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	brokerName := c.String("b")
	serviceName := c.String("e")
	orgName := c.String("o")

	if brokerName != "" && serviceName != "" && orgName != "" {
		cmd.ui.Say(T("Getting service access for broker {{.Broker}} and service {{.Service}} and organization {{.Organization}} as {{.Username}}...", map[string]interface{}{
			"Broker":       terminal.EntityNameColor(brokerName),
			"Service":      terminal.EntityNameColor(serviceName),
			"Organization": terminal.EntityNameColor(orgName),
			"Username":     terminal.EntityNameColor(cmd.config.Username())}))
	} else if serviceName != "" && orgName != "" {
		cmd.ui.Say(T("Getting service access for service {{.Service}} and organization {{.Organization}} as {{.Username}}...", map[string]interface{}{
			"Service":      terminal.EntityNameColor(serviceName),
			"Organization": terminal.EntityNameColor(orgName),
			"Username":     terminal.EntityNameColor(cmd.config.Username())}))
	} else if brokerName != "" && orgName != "" {
		cmd.ui.Say(T("Getting service access for broker {{.Broker}} and organization {{.Organization}} as {{.Username}}...", map[string]interface{}{
			"Broker":       terminal.EntityNameColor(brokerName),
			"Organization": terminal.EntityNameColor(orgName),
			"Username":     terminal.EntityNameColor(cmd.config.Username())}))
	} else if brokerName != "" && serviceName != "" {
		cmd.ui.Say(T("Getting service access for broker {{.Broker}} and service {{.Service}} as {{.Username}}...", map[string]interface{}{
			"Broker":   terminal.EntityNameColor(brokerName),
			"Service":  terminal.EntityNameColor(serviceName),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))
	} else if brokerName != "" {
		cmd.ui.Say(T("Getting service access for broker {{.Broker}} as {{.Username}}...", map[string]interface{}{
			"Broker":   terminal.EntityNameColor(brokerName),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))
	} else if serviceName != "" {
		cmd.ui.Say(T("Getting service access for service {{.Service}} as {{.Username}}...", map[string]interface{}{
			"Service":  terminal.EntityNameColor(serviceName),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))
	} else if orgName != "" {
		cmd.ui.Say(T("Getting service access for organization {{.Organization}} as {{.Username}}...", map[string]interface{}{
			"Organization": terminal.EntityNameColor(orgName),
			"Username":     terminal.EntityNameColor(cmd.config.Username())}))
	} else {
		cmd.ui.Say(T("Getting service access as {{.Username}}...", map[string]interface{}{
			"Username": terminal.EntityNameColor(cmd.config.Username())}))
	}

	brokers, err := cmd.actor.FilterBrokers(brokerName, serviceName, orgName)
	if err != nil {
		cmd.ui.Failed(T("Failed fetching service brokers.\n{{.Error}}", map[string]interface{}{"Error": err}))
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
