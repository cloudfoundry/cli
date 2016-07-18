package organization

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DeleteOrg struct {
	ui      terminal.UI
	config  coreconfig.ReadWriter
	orgRepo organizations.OrganizationRepository
	orgReq  requirements.OrganizationRequirement
}

func init() {
	commandregistry.Register(&DeleteOrg{})
}

func (cmd *DeleteOrg) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-org",
		Description: T("Delete an org"),
		Usage: []string{
			T("CF_NAME delete-org ORG [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteOrg) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-org"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *DeleteOrg) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	return cmd
}

func (cmd *DeleteOrg) Execute(c flags.FlagContext) error {
	orgName := c.Args()[0]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDeleteWithAssociations(T("org"), orgName) {
			return nil
		}
	}

	cmd.ui.Say(T("Deleting org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(orgName),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	org, err := cmd.orgRepo.FindByName(orgName)

	switch err.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Org {{.OrgName}} does not exist.",
			map[string]interface{}{"OrgName": orgName}))
		return nil
	default:
		return err
	}

	err = cmd.orgRepo.Delete(org.GUID)
	if err != nil {
		return err
	}

	if org.GUID == cmd.config.OrganizationFields().GUID {
		cmd.config.SetOrganizationFields(models.OrganizationFields{})
		cmd.config.SetSpaceFields(models.SpaceFields{})
	}

	cmd.ui.Ok()
	return nil
}
