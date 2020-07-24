package domain

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
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

func (cmd *DeleteDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-domain"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	loginReq := requirementsFactory.NewLoginRequirement()
	cmd.orgReq = requirementsFactory.NewTargetedOrgRequirement()

	reqs := []requirements.Requirement{
		loginReq,
		cmd.orgReq,
	}

	return reqs, nil
}

func (cmd *DeleteDomain) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *DeleteDomain) Execute(c flags.FlagContext) error {
	domainName := c.Args()[0]
	domain, err := cmd.domainRepo.FindByNameInOrg(domainName, cmd.orgReq.GetOrganizationFields().GUID)

	switch err.(type) {
	case nil:
		if domain.Shared {
			return errors.New(T("domain {{.DomainName}} is a shared domain, not an owned domain.",
				map[string]interface{}{
					"DomainName": domainName}))
		}
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(err.Error())
		return nil
	default:
		return errors.New(T("Error finding domain {{.DomainName}}\n{{.APIErr}}",
			map[string]interface{}{"DomainName": domainName, "APIErr": err.Error()}))
	}

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("domain"), domainName) {
			return nil
		}
	}

	cmd.ui.Say(T("Deleting domain {{.DomainName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domainName),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	err = cmd.domainRepo.Delete(domain.GUID)
	if err != nil {
		return errors.New(T("Error deleting domain {{.DomainName}}\n{{.APIErr}}",
			map[string]interface{}{"DomainName": domainName, "APIErr": err.Error()}))
	}

	cmd.ui.Ok()
	return nil
}
