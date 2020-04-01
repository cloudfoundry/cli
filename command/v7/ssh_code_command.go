package v7

type SSHCodeCommand struct {
	BaseCommand
	usage           interface{} `usage:"CF_NAME ssh-code"`
	relatedCommands interface{} `related_commands:"curl, ssh"`
}

func (cmd SSHCodeCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	code, err := cmd.Actor.GetSSHPasscode()
	cmd.UI.DisplayText("{{.SSHCode}}", map[string]interface{}{"SSHCode": code})
	return err
}
