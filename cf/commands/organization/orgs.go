package organization

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin/models"
)

const orgLimit = 0

type ListOrgs struct {
	ui              terminal.UI
	config          coreconfig.Reader
	orgRepo         organizations.OrganizationRepository
	pluginOrgsModel *[]plugin_models.GetOrgs_Model
	pluginCall      bool
}

func init() {
	commandregistry.Register(&ListOrgs{})
}

func (cmd *ListOrgs) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "orgs",
		ShortName:   "o",
		Description: T("List all orgs"),
		Usage: []string{
			"CF_NAME orgs",
		},
	}
}

func (cmd *ListOrgs) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
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

func (cmd *ListOrgs) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.pluginOrgsModel = deps.PluginModels.Organizations
	cmd.pluginCall = pluginCall
	return cmd
}

func (cmd ListOrgs) Execute(fc flags.FlagContext) error {
	cmd.ui.Say(T("Getting orgs as {{.Username}}...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	noOrgs := true
	table := cmd.ui.Table([]string{T("name")})

	orgs, err := cmd.orgRepo.ListOrgs(orgLimit)
	if err != nil {
		return err
	}
	for _, org := range orgs {
		table.Add(org.Name)
		noOrgs = false
	}

	table.Print()

	if err != nil {
		return errors.New(T("Failed fetching orgs.\n{{.APIErr}}",
			map[string]interface{}{"APIErr": err}))
	}

	if noOrgs {
		cmd.ui.Say(T("No orgs found"))
	}

	if cmd.pluginCall {
		cmd.populatePluginModel(orgs)
	}
	return nil
}

func (cmd *ListOrgs) populatePluginModel(orgs []models.Organization) {
	for _, org := range orgs {
		orgModel := plugin_models.GetOrgs_Model{}
		orgModel.Name = org.Name
		orgModel.Guid = org.GUID
		*(cmd.pluginOrgsModel) = append(*(cmd.pluginOrgsModel), orgModel)
	}
}
