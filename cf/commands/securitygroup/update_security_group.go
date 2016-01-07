package securitygroup

import (
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/json"
)

type UpdateSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	configRepo        core_config.Reader
}

func init() {
	command_registry.Register(&UpdateSecurityGroup{})
}

func (cmd *UpdateSecurityGroup) MetaData() command_registry.CommandMetadata {
	primaryUsage := T("CF_NAME update-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE")
	secondaryUsage := T("   The provided path can be an absolute or relative path to a file.\n   It should have a single array with JSON objects inside describing the rules.")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return command_registry.CommandMetadata{
		Name:        "update-security-group",
		Description: T("Update a security group"),
		Usage:       strings.Join([]string{primaryUsage, secondaryUsage, tipUsage}, "\n\n"),
	}
}

func (cmd *UpdateSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SECURITY_GROUP and PATH_TO_JSON_RULES_FILE as arguments\n\n") + command_registry.Commands.CommandUsage("update-security-group"))
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd *UpdateSecurityGroup) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	return cmd
}

func (cmd *UpdateSecurityGroup) Execute(context flags.FlagContext) {
	name := context.Args()[0]
	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	pathToJSONFile := context.Args()[1]
	rules, err := json.ParseJsonArray(pathToJSONFile)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say(T("Updating security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))
	err = cmd.securityGroupRepo.Update(securityGroup.Guid, rules)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
}
