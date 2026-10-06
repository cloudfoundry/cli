package v7

import "errors"

type ServiceAccountsCommand struct {
	BaseCommand
	usage           interface{} `usage:"CF_NAME service-accounts"`
	relatedCommands interface{} `related_commands:"create-service-account"`
}

func (cmd ServiceAccountsCommand) Execute(args []string) error {
	return errors.New("service account listing is not implemented")
}
