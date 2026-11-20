package v7

type CreateServiceAccountCommand struct {
	BaseCommand
	RequiredArgs struct {
		Name string `positional-arg-name:"SERVICE_ACCOUNT_NAME" required:"yes"`
	} `positional-args:"yes"`
	Description string      `long:"description" description:"Description of the service account"`
	usage       interface{} `usage:"CF_NAME create-service-account SERVICE_ACCOUNT_NAME [--description DESCRIPTION]"`
}

func (cmd CreateServiceAccountCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor("Creating service account {{.Name}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
		"Name":  cmd.RequiredArgs.Name,
		"Org":   cmd.Config.TargetedOrganization().Name,
		"Space": cmd.Config.TargetedSpace().Name,
		"User":  user.Name,
	})
	account, warnings, err := cmd.Actor.CreateServiceAccountInSpace(cmd.RequiredArgs.Name, cmd.Description, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Client ID: {{.ClientID}}\nCertificate DNS SAN: {{.SAN}}\nStatus: {{.Status}}", map[string]interface{}{
		"ClientID": account.ClientID, "SAN": account.CertificateDNSSAN, "Status": account.Status,
	})
	cmd.UI.DisplayText("Provisioning occurs on first bind. Account roles must be granted explicitly.")
	return nil
}
