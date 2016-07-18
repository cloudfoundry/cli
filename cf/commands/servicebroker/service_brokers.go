package servicebroker

import (
	"sort"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListServiceBrokers struct {
	ui     terminal.UI
	config coreconfig.Reader
	repo   api.ServiceBrokerRepository
}

type serviceBrokerTable []serviceBrokerRow

type serviceBrokerRow struct {
	name string
	url  string
}

func init() {
	commandregistry.Register(&ListServiceBrokers{})
}

func (cmd *ListServiceBrokers) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "service-brokers",
		Description: T("List service brokers"),
		Usage: []string{
			"CF_NAME service-brokers",
		},
	}
}

func (cmd *ListServiceBrokers) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
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

	return reqs
}

func (cmd *ListServiceBrokers) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.repo = deps.RepoLocator.GetServiceBrokerRepository()
	return cmd
}

func (cmd *ListServiceBrokers) Execute(c flags.FlagContext) error {
	sbTable := serviceBrokerTable{}

	cmd.ui.Say(T("Getting service brokers as {{.Username}}...\n",
		map[string]interface{}{
			"Username": terminal.EntityNameColor(cmd.config.Username()),
		}))

	table := cmd.ui.Table([]string{T("name"), T("url")})
	foundBrokers := false
	err := cmd.repo.ListServiceBrokers(func(serviceBroker models.ServiceBroker) bool {
		sbTable = append(sbTable, serviceBrokerRow{
			name: serviceBroker.Name,
			url:  serviceBroker.URL,
		})
		foundBrokers = true
		return true
	})

	sort.Sort(sbTable)

	for _, sb := range sbTable {
		table.Add(sb.name, sb.url)
	}

	table.Print()

	if err != nil {
		return err
	}

	if !foundBrokers {
		cmd.ui.Say(T("No service brokers found"))
	}
	return nil
}

func (a serviceBrokerTable) Len() int           { return len(a) }
func (a serviceBrokerTable) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a serviceBrokerTable) Less(i, j int) bool { return a[i].name < a[j].name }
