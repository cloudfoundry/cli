package v7

import "errors"

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
	return errors.New("service account command is not implemented")
}
func (cmd DeleteServiceAccountCommand) Execute(args []string) error {
	return errors.New("service account command is not implemented")
}
func (cmd EnableServiceAccountCommand) Execute(args []string) error {
	return errors.New("service account command is not implemented")
}
func (cmd DisableServiceAccountCommand) Execute(args []string) error {
	return errors.New("service account command is not implemented")
}
func (cmd BindServiceAccountCommand) Execute(args []string) error {
	return errors.New("service account command is not implemented")
}
func (cmd UnbindServiceAccountCommand) Execute(args []string) error {
	return errors.New("service account command is not implemented")
}
