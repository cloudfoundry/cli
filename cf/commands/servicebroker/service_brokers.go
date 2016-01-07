package servicebroker

import (
	"sort"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type ListServiceBrokers struct {
	ui     terminal.UI
	config core_config.Reader
	repo   api.ServiceBrokerRepository
}

type serviceBrokerTable []serviceBrokerRow

type serviceBrokerRow struct {
	name string
	url  string
}

func init() {
	command_registry.Register(&ListServiceBrokers{})
}

func (cmd *ListServiceBrokers) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "service-brokers",
		Description: T("List service brokers"),
		Usage:       "CF_NAME service-brokers",
	}
}

func (cmd *ListServiceBrokers) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("service-brokers"))
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd *ListServiceBrokers) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.repo = deps.RepoLocator.GetServiceBrokerRepository()
	return cmd
}

func (cmd *ListServiceBrokers) Execute(c flags.FlagContext) {
	sbTable := serviceBrokerTable{}

	cmd.ui.Say(T("Getting service brokers as {{.Username}}...\n",
		map[string]interface{}{
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	table := cmd.ui.Table([]string{T("name"), T("url")})
	foundBrokers := false
	apiErr := cmd.repo.ListServiceBrokers(func(serviceBroker models.ServiceBroker) bool {
		sbTable = append(sbTable, serviceBrokerRow{
			name: serviceBroker.Name,
			url:  serviceBroker.Url,
		})
		foundBrokers = true
		return true
	})

	sort.Sort(sbTable)

	for _, sb := range sbTable {
		table.Add(sb.name, sb.url)
	}

	table.Print()

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	if !foundBrokers {
		cmd.ui.Say(T("No service brokers found"))
	}
}

func (a serviceBrokerTable) Len() int           { return len(a) }
func (a serviceBrokerTable) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a serviceBrokerTable) Less(i, j int) bool { return a[i].name < a[j].name }
