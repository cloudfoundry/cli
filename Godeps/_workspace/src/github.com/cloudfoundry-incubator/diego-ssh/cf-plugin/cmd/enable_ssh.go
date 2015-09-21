package cmd

import (
	"fmt"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/app"
)

const EnableSSHUsage = "cf enable-ssh APP_NAME"

func EnableSSH(args []string, appFactory app.AppFactory) error {
	if len(args) != 2 || args[0] != "enable-ssh" {
		return fmt.Errorf("%s\n%s", "Invalid usage", EnableSSHUsage)
	}

	app, err := appFactory.Get(args[1])
	if err != nil {
		return err
	}

	err = appFactory.SetBool(app, "enable_ssh", true)
	if err != nil {
		return err
	}

	return nil
}
