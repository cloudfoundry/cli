package securitygroup

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/api/securitygroups"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/util/json"
)

type UpdateSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo securitygroups.SecurityGroupRepo
	configRepo        coreconfig.Reader
}

func init() {
	commandregistry.Register(&UpdateSecurityGroup{})
}

func (cmd *UpdateSecurityGroup) MetaData() commandregistry.CommandMetadata {
	primaryUsage := T("CF_NAME update-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE")
	secondaryUsage := T("   The provided path can be an absolute or relative path to a file.\n   It should have a single array with JSON objects inside describing the rules.")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return commandregistry.CommandMetadata{
		Name:        "update-security-group",
		Description: T("Update a security group"),
		Usage: []string{
			primaryUsage,
			"\n\n",
			secondaryUsage,
			"\n\n",
			tipUsage,
		},
	}
}

func (cmd *UpdateSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SECURITY_GROUP and PATH_TO_JSON_RULES_FILE as arguments\n\n") + commandregistry.Commands.CommandUsage("update-security-group"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	reqs := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return reqs, nil
}

func (cmd *UpdateSecurityGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	return cmd
}

func (cmd *UpdateSecurityGroup) Execute(context flags.FlagContext) error {
	name := context.Args()[0]
	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		return err
	}

	pathToJSONFile := context.Args()[1]
	rules, err := json.ParseJSONArray(pathToJSONFile)
	if err != nil {
		return err
	}

	cmd.ui.Say(T("Updating security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))
	err = cmd.securityGroupRepo.Update(securityGroup.GUID, rules)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
	return nil
}
