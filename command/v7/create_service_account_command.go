package v7

import "errors"

type CreateServiceAccountCommand struct {
	BaseCommand
	RequiredArgs struct {
		Name string `positional-arg-name:"SERVICE_ACCOUNT_NAME" required:"yes"`
	} `positional-args:"yes"`
	Description string      `long:"description" description:"Description of the service account"`
	usage       interface{} `usage:"CF_NAME create-service-account SERVICE_ACCOUNT_NAME [--description DESCRIPTION]"`
}

func (cmd CreateServiceAccountCommand) Execute(args []string) error {
	return errors.New("service account creation is not implemented")
}
