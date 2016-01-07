package organization

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/models"
)

type ListOrgs struct {
	ui              terminal.UI
	config          core_config.Reader
	orgRepo         organizations.OrganizationRepository
	pluginOrgsModel *[]plugin_models.GetOrgs_Model
	pluginCall      bool
}

func init() {
	command_registry.Register(&ListOrgs{})
}

func (cmd *ListOrgs) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "orgs",
		ShortName:   "o",
		Description: T("List all orgs"),
		Usage:       "CF_NAME orgs",
	}
}

func (cmd *ListOrgs) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("orgs"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *ListOrgs) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.pluginOrgsModel = deps.PluginModels.Organizations
	cmd.pluginCall = pluginCall
	return cmd
}

func (cmd ListOrgs) Execute(fc flags.FlagContext) {
	cmd.ui.Say(T("Getting orgs as {{.Username}}...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	noOrgs := true
	table := cmd.ui.Table([]string{T("name")})

	orgs, apiErr := cmd.orgRepo.ListOrgs()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}
	for _, org := range orgs {
		table.Add(org.Name)
		noOrgs = false
	}

	table.Print()

	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching orgs.\n{{.ApiErr}}",
			map[string]interface{}{"ApiErr": apiErr}))
		return
	}

	if noOrgs {
		cmd.ui.Say(T("No orgs found"))
	}

	if cmd.pluginCall {
		cmd.populatePluginModel(orgs)
	}

}

func (cmd *ListOrgs) populatePluginModel(orgs []models.Organization) {
	for _, org := range orgs {
		orgModel := plugin_models.GetOrgs_Model{}
		orgModel.Name = org.Name
		orgModel.Guid = org.Guid
		*(cmd.pluginOrgsModel) = append(*(cmd.pluginOrgsModel), orgModel)
	}
}
