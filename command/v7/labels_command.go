package v7

import (
	"code.cloudfoundry.org/cli/command"
)

type LabelsCommand struct {
	usage interface{} `usage:"CF_NAME labels RESOURCE RESOURCE_NAME\n\n EXAMPLES:\n   cf labels app dora \n\nRESOURCES:\n   APP\n\nSEE ALSO:\n   set-label, delete-label"`
}

func (cmd *LabelsCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (cmd LabelsCommand) Execute(args []string) error {
	return nil
}
