package cmd

import (
	"fmt"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/app"
)

const DisableSSHUsage = "cf disable-ssh APP_NAME"

func DisableSSH(args []string, appFactory app.AppFactory) error {
	if len(args) != 2 || args[0] != "disable-ssh" {
		return fmt.Errorf("%s\n%s", "Invalid usage", DisableSSHUsage)
	}

	app, err := appFactory.Get(args[1])
	if err != nil {
		return err
	}

	err = appFactory.SetBool(app, "enable_ssh", false)
	if err != nil {
		return err
	}

	return nil
}
