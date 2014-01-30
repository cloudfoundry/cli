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
	"runtime/debug"
	"strings"
)

func main() {
	defer func() {
		err := recover()
		if err != nil {
			switch err := err.(type) {
			case error:
				displayCrashDialog(err.Error())
			case string:
				displayCrashDialog(err)
			default:
				displayCrashDialog("An unexpected type of error")
			}
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

	args := os.Args
	if len(args) == 2 && args[1][0] == '-' && args[1] != "-v" && args[1] != "--version" {
		args[1] = "help"
	}

	app.Run(args)
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
   CF_COLOR=false - will not colorize output
   CF_HOME=path/to/config/ override default config directory
   CF_STAGING_TIMEOUT=15 max wait time for buildpack staging, in minutes
   CF_STARTUP_TIMEOUT=5 max wait time for app instance startup, in minutes
   CF_TRACE=true - print API request diagnostics to stdout
   CF_TRACE=path/to/trace.log - append API request diagnostics to a log file
   HTTP_PROXY=http://proxy.example.com:8080 - enable HTTP proxying for API requests
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
		termUI.Failed(fmt.Sprintf("Error loading config file: %s",err))
		configRepo.Delete()
		os.Exit(1)
		return
	}
	return
}

func displayCrashDialog(errorMessage string) {
	formattedString := `

Aww shucks.

Something completely unexpected happened. This is a bug in %s.
Please file this bug : https://github.com/cloudfoundry/cli/issues
Tell us that you ran this command:

	%s

this error occurred:

	%s

and this stack trace:

%s
	`

	stackTrace := "\t" + strings.Replace(string(debug.Stack()), "\n", "\n\t", -1)
	println(fmt.Sprintf(formattedString, cf.Name(), strings.Join(os.Args, " "), errorMessage, stackTrace))
	os.Exit(1)
}
