package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteLabelCommand struct {
	RequiredArgs flag.DeleteLabelArgs `positional-args:"yes"`
	usage        interface{}          `usage:"cf delete-label RESOURCE RESOURCE_NAME KEY\n\n EXAMPLES:\n   cf delete-label app dora ci_signature_sha2\n\nRESOURCES:\n   APP\n\nSEE ALSO:\n   set-label, labels"`
}
