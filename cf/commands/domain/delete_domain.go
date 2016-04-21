package domain

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type DeleteDomain struct {
	ui         terminal.UI
	config     coreconfig.Reader
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func init() {
	commandregistry.Register(&DeleteDomain{})
}

func (cmd *DeleteDomain) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-domain",
		Description: T("Delete a domain"),
		Usage: []string{
			T("CF_NAME delete-domain DOMAIN [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-domain"))
	}

	loginReq := requirementsFactory.NewLoginRequirement()
	cmd.orgReq = requirementsFactory.NewTargetedOrgRequirement()

	reqs := []requirements.Requirement{
		loginReq,
		cmd.orgReq,
	}

	return reqs
}

func (cmd *DeleteDomain) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *DeleteDomain) Execute(c flags.FlagContext) {
	domainName := c.Args()[0]
	domain, apiErr := cmd.domainRepo.FindByNameInOrg(domainName, cmd.orgReq.GetOrganizationFields().GUID)

	switch apiErr.(type) {
	case nil:
		if domain.Shared {
			cmd.ui.Failed(T("domain {{.DomainName}} is a shared domain, not an owned domain.",
				map[string]interface{}{
					"DomainName": domainName}))
			return
		}
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(apiErr.Error())
		return
	default:
		cmd.ui.Failed(T("Error finding domain {{.DomainName}}\n{{.APIErr}}",
			map[string]interface{}{"DomainName": domainName, "APIErr": apiErr.Error()}))
		return
	}

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("domain"), domainName) {
			return
		}
	}

	cmd.ui.Say(T("Deleting domain {{.DomainName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domainName),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	apiErr = cmd.domainRepo.Delete(domain.GUID)
	if apiErr != nil {
		cmd.ui.Failed(T("Error deleting domain {{.DomainName}}\n{{.APIErr}}",
			map[string]interface{}{"DomainName": domainName, "APIErr": apiErr.Error()}))
		return
	}

	cmd.ui.Ok()
}
