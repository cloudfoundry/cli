package cmd

import (
	"fmt"
	"io"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/space"
)

const SSHAllowedUsage = "cf space-ssh-allowed SPACE_NAME"

func SSHAllowed(args []string, spaceFactory space.SpaceFactory, output io.Writer) error {
	if len(args) != 2 || args[0] != "space-ssh-allowed" {
		return fmt.Errorf("%s\n%s", "Invalid usage", SSHAllowedUsage)
	}

	space, err := spaceFactory.Get(args[1])
	if err != nil {
		return err
	}

	fmt.Fprintf(output, "%t", space.AllowSSH)
	return nil
}
