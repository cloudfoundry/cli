package main

import (
	"os"
	"cf/app"
	"cf/requirements"
	"cf/commands"
	"cf/api"
	"cf"
	"fmt"
	"cf/terminal"
	"cf/configuration"
	"github.com/codegangsta/cli"
	"cf/net"
)

func main() {

	if os.Getenv("CF_COLOR") == "" {
		os.Setenv("CF_COLOR", "true")
	}

	termUI := terminal.NewUI()
	assignTemplates()
	configRepo := configuration.NewConfigurationDiskRepository()
	config := loadConfig(termUI, configRepo)

	repoLocator := api.NewRepositoryLocator(config, configRepo, map[string]net.Gateway{
		"auth": net.NewUAAGateway(),
		"cloud-controller": net.NewCloudControllerGateway(),
		"uaa": net.NewUAAGateway(),
	})

	cmdFactory := commands.NewFactory(termUI, config, configRepo, repoLocator)
	reqFactory := requirements.NewFactory(termUI, config, repoLocator)
	cmdRunner := commands.NewRunner(cmdFactory, reqFactory)

	app, err := app.NewApp(cmdRunner)
	if err != nil {
		return
	}
	app.Run(os.Args)
}

func assignTemplates() {
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   [environment variables] {{.Name}} [global options] command [arguments...] [command options]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Description}}
   {{end}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
ENVIRONMENT VARIABLES:
   CF_TRACE=true - will output HTTP requests and responses during command
   HTTP_PROXY=http://proxy.example.com:8080 - set to your proxy
`

	cli.CommandHelpTemplate = `NAME:
   {{.Name}} - {{.Description}}
{{with .ShortName}}
ALIAS:
   {{.}}
{{end}}
USAGE:
   {{.Usage}}{{with .Flags}}

OPTIONS:
   {{range .}}{{.}}
   {{end}}{{else}}
{{end}}`

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
