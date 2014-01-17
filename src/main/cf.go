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
	"github.com/codegangsta/cli"
	"os"
)

func main() {
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

func init() {
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
