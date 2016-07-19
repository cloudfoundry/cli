package serviceaccess

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/actors"
	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type ServiceAccess struct {
	ui             terminal.UI
	config         coreconfig.Reader
	actor          actors.ServiceActor
	tokenRefresher authentication.TokenRefresher
}

func init() {
	commandregistry.Register(&ServiceAccess{})
}

func (cmd *ServiceAccess) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["b"] = &flags.StringFlag{ShortName: "b", Usage: T("Access for plans of a particular broker")}
	fs["e"] = &flags.StringFlag{ShortName: "e", Usage: T("Access for service name of a particular service offering")}
	fs["o"] = &flags.StringFlag{ShortName: "o", Usage: T("Plans accessible by a particular organization")}

	return commandregistry.CommandMetadata{
		Name:        "service-access",
		Description: T("List service access settings"),
		Usage: []string{
			"CF_NAME service-access [-b BROKER] [-e SERVICE] [-o ORG]",
		},
		Flags: fs,
	}
}

func (cmd *ServiceAccess) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *ServiceAccess) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.actor = deps.ServiceHandler
	cmd.tokenRefresher = deps.RepoLocator.GetAuthenticationRepository()
	return cmd
}

func (cmd *ServiceAccess) Execute(c flags.FlagContext) error {
	_, err := cmd.tokenRefresher.RefreshAuthToken()
	if err != nil {
		return err
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
		return err
	}
	cmd.printTable(brokers)
	return nil
}

func (cmd ServiceAccess) printTable(brokers []models.ServiceBroker) error {
	for _, serviceBroker := range brokers {
		cmd.ui.Say(fmt.Sprintf(T("broker: {{.Name}}", map[string]interface{}{"Name": serviceBroker.Name})))

		table := cmd.ui.Table([]string{"", T("service"), T("plan"), T("access"), T("orgs")})
		for _, service := range serviceBroker.Services {
			if len(service.Plans) > 0 {
				for _, plan := range service.Plans {
					table.Add("", service.Label, plan.Name, cmd.formatAccess(plan.Public, plan.OrgNames), strings.Join(plan.OrgNames, ","))
				}
			} else {
				table.Add("", service.Label, "", "", "")
			}
		}
		err := table.Print()
		if err != nil {
			return err
		}

		cmd.ui.Say("")
	}
	return nil
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
