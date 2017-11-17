package actionerror

import (
	"fmt"
	"strings"
)

type CommandLineOptionsAndManifestConflictError struct {
	ManifestAttribute  string
	CommandLineOptions []string
}

func (e CommandLineOptionsAndManifestConflictError) Error() string {
	return fmt.Sprintf("cannot use manifest attribute %s with command line options: %s",
		e.ManifestAttribute, strings.Join(e.CommandLineOptions, ", "),
	)
}
