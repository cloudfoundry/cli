package v7

import "code.cloudfoundry.org/cli/command"

type SetLabelCommand struct {
	usage       interface{} `usage:"cf set-label RESOURCE RESOURCE_NAME KEY=VALUE...\n\nEXAMPLES:\n   cf set-label app dora env=production\n\n RESOURCES:\n   APP\n\nSEE ALSO:\n   delete-label, labels"`
	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetHealthCheckActor
}
