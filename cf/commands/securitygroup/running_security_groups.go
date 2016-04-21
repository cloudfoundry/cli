package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/securitygroups/defaults/running"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type listRunningSecurityGroups struct {
	ui                       terminal.UI
	runningSecurityGroupRepo running.RunningSecurityGroupsRepo
	configRepo               coreconfig.Reader
}

func init() {
	commandregistry.Register(&listRunningSecurityGroups{})
}

func (cmd *listRunningSecurityGroups) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "running-security-groups",
		Description: T("List security groups in the set of security groups for running applications"),
		Usage: []string{
			"CF_NAME running-security-groups",
		},
	}
}

func (cmd *listRunningSecurityGroups) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
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

func (cmd *listRunningSecurityGroups) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.runningSecurityGroupRepo = deps.RepoLocator.GetRunningSecurityGroupsRepository()
	return cmd
}

func (cmd *listRunningSecurityGroups) Execute(context flags.FlagContext) {
	cmd.ui.Say(T("Acquiring running security groups as '{{.username}}'", map[string]interface{}{
		"username": terminal.EntityNameColor(cmd.configRepo.Username()),
	}))

	defaultSecurityGroupsFields, err := cmd.runningSecurityGroupRepo.List()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(defaultSecurityGroupsFields) > 0 {
		for _, value := range defaultSecurityGroupsFields {
			cmd.ui.Say(value.Name)
		}
	} else {
		cmd.ui.Say(T("No running security groups set"))
	}
}
