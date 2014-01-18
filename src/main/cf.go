package main

import (
	"cf"
	"cf/api"
	"cf/app"
	"cf/commands"
	"cf/configuration"
	"cf/manifest"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"fileutils"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

func main() {
	defer func() {
		maybeSomething := recover()

		if maybeSomething != nil {
			displayCrashDialog()
		}
	}()

	fileutils.SetTmpPathPrefix("cf")

	if os.Getenv("CF_COLOR") == "" {
		os.Setenv("CF_COLOR", "true")
	}

	termUI := terminal.NewUI()
	configRepo := configuration.NewConfigurationDiskRepository()
	config := loadConfig(termUI, configRepo)
	manifestRepo := manifest.NewManifestDiskRepository()
	repoLocator := api.NewRepositoryLocator(config, configRepo, map[string]net.Gateway{
		"auth":             net.NewUAAGateway(),
		"cloud-controller": net.NewCloudControllerGateway(),
		"uaa":              net.NewUAAGateway(),
	})

	cmdFactory := commands.NewFactory(termUI, config, configRepo, manifestRepo, repoLocator)
	reqFactory := requirements.NewFactory(termUI, config, repoLocator)
	cmdRunner := commands.NewRunner(cmdFactory, reqFactory)

	app, err := app.NewApp(cmdRunner)
	if err != nil {
		return
	}
	app.Run(os.Args)
}

func loadConfig(termUI terminal.UI, configRepo configuration.ConfigurationRepository) (config *configuration.Configuration) {
	config, err := configRepo.Get()
	if err != nil {
		termUI.Failed(fmt.Sprintf(
			"Error loading config. Please reset target (%s) and log in (%s).",
			terminal.CommandColor(fmt.Sprintf("%s target", cf.Name())),
			terminal.CommandColor(fmt.Sprintf("%s login", cf.Name())),
		))
		configRepo.Delete()
		os.Exit(1)
		return
	}
	return
}

func displayCrashDialog() {
	formattedString := `

       d8888 888       888 888       888        .d8888b.  888    888 888     888  .d8888b.  888    d8P   .d8888b.
      d88888 888   o   888 888   o   888       d88P  Y88b 888    888 888     888 d88P  Y88b 888   d8P   d88P  Y88b
     d88P888 888  d8b  888 888  d8b  888       Y88b.      888    888 888     888 888    888 888  d8P    Y88b.
    d88P 888 888 d888b 888 888 d888b 888        "Y888b.   8888888888 888     888 888        888d88K      "Y888b.
   d88P  888 888d88888b888 888d88888b888           "Y88b. 888    888 888     888 888        8888888b        "Y88b.
  d88P   888 88888P Y88888 88888P Y88888             "888 888    888 888     888 888    888 888  Y88b         "888
 d8888888888 8888P   Y8888 8888P   Y8888       Y88b  d88P 888    888 Y88b. .d88P Y88b  d88P 888   Y88b  Y88b  d88P d8b
d88P     888 888P     Y888 888P     Y888        "Y8888P"  888    888  "Y88888P"   "Y8888P"  888    Y88b  "Y8888P"  Y8P


Something completely unexpected happened. This is a bug in %s.
Please file this bug : https://github.com/cloudfoundry/cli/issues
Tell us that you ran this command:

	%s

and got this stack trace:

%s
	`

	stackTrace := "\t" + strings.Replace(string(debug.Stack()), "\n", "\n\t", -1)
	println(fmt.Sprintf(formattedString, cf.Name(), strings.Join(os.Args, " "), stackTrace))
	os.Exit(1)
}
