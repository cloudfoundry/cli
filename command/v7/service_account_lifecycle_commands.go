package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/util/ui"
)

type serviceAccountNameArgs struct {
	Name string `positional-arg-name:"SERVICE_ACCOUNT_NAME" required:"yes"`
}

type ServiceAccountCommand struct {
	BaseCommand
	RequiredArgs    serviceAccountNameArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME service-account SERVICE_ACCOUNT_NAME"`
	relatedCommands interface{}            `related_commands:"service-accounts, bind-service-account, enable-service-account, disable-service-account"`
}

type DeleteServiceAccountCommand struct {
	BaseCommand
	RequiredArgs    serviceAccountNameArgs `positional-args:"yes"`
	Force           bool                   `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}            `usage:"CF_NAME delete-service-account SERVICE_ACCOUNT_NAME [-f]"`
	relatedCommands interface{}            `related_commands:"service-accounts, unbind-service-account"`
}

type EnableServiceAccountCommand struct {
	BaseCommand
	RequiredArgs    serviceAccountNameArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME enable-service-account SERVICE_ACCOUNT_NAME"`
	relatedCommands interface{}            `related_commands:"service-account, disable-service-account"`
}

type DisableServiceAccountCommand struct {
	BaseCommand
	RequiredArgs    serviceAccountNameArgs `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME disable-service-account SERVICE_ACCOUNT_NAME"`
	relatedCommands interface{}            `related_commands:"service-account, enable-service-account"`
}

type BindServiceAccountCommand struct {
	BaseCommand
	RequiredArgs struct {
		AppName     string `positional-arg-name:"APP_NAME" required:"yes"`
		AccountName string `positional-arg-name:"SERVICE_ACCOUNT_NAME" required:"yes"`
	} `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME bind-service-account APP_NAME SERVICE_ACCOUNT_NAME"`
	relatedCommands interface{} `related_commands:"service-accounts, unbind-service-account, restart"`
}

type UnbindServiceAccountCommand struct {
	BaseCommand
	RequiredArgs struct {
		AppName string `positional-arg-name:"APP_NAME" required:"yes"`
	} `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME unbind-service-account APP_NAME"`
	relatedCommands interface{} `related_commands:"bind-service-account, restart"`
}

func (cmd ServiceAccountCommand) Execute(args []string) error {
	if err := cmd.serviceAccountIntro("Getting service account", cmd.RequiredArgs.Name); err != nil {
		return err
	}
	account, warnings, err := cmd.Actor.GetServiceAccountByNameAndSpace(cmd.RequiredArgs.Name, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTableWithHeader("", [][]string{
		{cmd.UI.TranslateText("name:"), account.Name},
		{cmd.UI.TranslateText("guid:"), account.GUID},
		{cmd.UI.TranslateText("description:"), account.Description},
		{cmd.UI.TranslateText("enabled:"), strconv.FormatBool(account.Enabled)},
		{cmd.UI.TranslateText("status:"), account.Status},
		{cmd.UI.TranslateText("client id:"), account.ClientID},
		{cmd.UI.TranslateText("certificate DNS SAN:"), account.CertificateDNSSAN},
	}, ui.DefaultTableSpacePadding)
	return nil
}
func (cmd DeleteServiceAccountCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}
	if !cmd.Force {
		confirmed, err := cmd.UI.DisplayBoolPrompt(false, "Really delete service account {{.Name}}? Its name will remain permanently reserved.", map[string]interface{}{"Name": cmd.RequiredArgs.Name})
		if err != nil {
			return err
		}
		if !confirmed {
			cmd.UI.DisplayText("Delete cancelled")
			return nil
		}
	}
	if err := cmd.serviceAccountIntro("Deleting service account", cmd.RequiredArgs.Name); err != nil {
		return err
	}
	warnings, err := cmd.Actor.DeleteServiceAccountByNameAndSpace(cmd.RequiredArgs.Name, cmd.Config.TargetedSpace().GUID)
	return cmd.serviceAccountResult(warnings, err, "The service-account name remains permanently reserved.")
}
func (cmd EnableServiceAccountCommand) Execute(args []string) error {
	if err := cmd.serviceAccountIntro("Enabling service account", cmd.RequiredArgs.Name); err != nil {
		return err
	}
	warnings, err := cmd.Actor.SetServiceAccountEnabledByNameAndSpace(cmd.RequiredArgs.Name, cmd.Config.TargetedSpace().GUID, true)
	return cmd.serviceAccountResult(warnings, err, "")
}
func (cmd DisableServiceAccountCommand) Execute(args []string) error {
	if err := cmd.serviceAccountIntro("Disabling service account", cmd.RequiredArgs.Name); err != nil {
		return err
	}
	warnings, err := cmd.Actor.SetServiceAccountEnabledByNameAndSpace(cmd.RequiredArgs.Name, cmd.Config.TargetedSpace().GUID, false)
	return cmd.serviceAccountResult(warnings, err, "New token issuance is disabled. Existing tokens and certificates are not revoked.")
}
func (cmd BindServiceAccountCommand) Execute(args []string) error {
	if err := cmd.serviceAccountIntro("Binding service account", cmd.RequiredArgs.AccountName); err != nil {
		return err
	}
	warnings, err := cmd.Actor.BindServiceAccountByNameAndSpace(cmd.RequiredArgs.AppName, cmd.RequiredArgs.AccountName, cmd.Config.TargetedSpace().GUID)
	if err := cmd.serviceAccountResult(warnings, err, "Roles are not granted by binding."); err != nil {
		return err
	}
	cmd.serviceAccountRestartGuidance(cmd.RequiredArgs.AppName)
	return nil
}
func (cmd UnbindServiceAccountCommand) Execute(args []string) error {
	if err := cmd.serviceAccountIntro("Unbinding service account from app", cmd.RequiredArgs.AppName); err != nil {
		return err
	}
	warnings, err := cmd.Actor.UnbindServiceAccountByAppNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	if err := cmd.serviceAccountResult(warnings, err, "Existing tokens and certificates are not revoked."); err != nil {
		return err
	}
	cmd.serviceAccountRestartGuidance(cmd.RequiredArgs.AppName)
	return nil
}

func (cmd BaseCommand) serviceAccountIntro(action, name string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor("{{.Action}} {{.Name}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
		"Action": cmd.UI.TranslateText(action), "Name": name, "Org": cmd.Config.TargetedOrganization().Name, "Space": cmd.Config.TargetedSpace().Name, "User": user.Name,
	})
	return nil
}

func (cmd BaseCommand) serviceAccountResult(warnings v7action.Warnings, err error, guidance string) error {
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()
	if guidance != "" {
		cmd.UI.DisplayText(guidance)
	}
	return nil
}

func (cmd BaseCommand) serviceAccountRestartGuidance(appName string) {
	cmd.UI.DisplayText("Restart {{.AppName}} for the assignment change to take effect. Use '{{.Binary}} restart {{.AppName}}'.", map[string]interface{}{"AppName": appName, "Binary": cmd.Config.BinaryName()})
}
