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
	"github.com/nu7hatch/gouuid"
)

type cliDependencies struct {
	termUI         terminal.UI
	configRepo     configuration.Repository
	manifestRepo   manifest.ManifestRepository
	apiRepoLocator api.RepositoryLocator
}

func setupDependencies() (deps *cliDependencies) {
	uuid, err := uuid.NewV4()
	if err != nil {
		deps.termUI.Failed(fmt.Sprintf("Failed to create uuid: %s", err))
	}
	fileutils.SetTmpPathPrefix(uuid.String())

	if os.Getenv("CF_COLOR") == "" {
		os.Setenv("CF_COLOR", "true")
	}

	deps = new(cliDependencies)

	deps.termUI = terminal.NewUI(os.Stdin)

	deps.manifestRepo = manifest.NewManifestDiskRepository()

	deps.configRepo = configuration.NewRepositoryFromFilepath(configuration.DefaultFilePath(), func(err error) {
		if err != nil {
			deps.termUI.Failed(fmt.Sprintf("Config error: %s", err))
		}
	})

	deps.apiRepoLocator = api.NewRepositoryLocator(deps.configRepo, map[string]net.Gateway{
		"auth":             net.NewUAAGateway(),
		"cloud-controller": net.NewCloudControllerGateway(),
		"uaa":              net.NewUAAGateway(),
	})

	return
}

func teardownDependencies(deps *cliDependencies) {
	deps.configRepo.Close()
}

func main() {
	defer handlePanics()

	deps := setupDependencies()
	defer teardownDependencies(deps)

	cmdFactory := commands.NewFactory(deps.termUI, deps.configRepo, deps.manifestRepo, deps.apiRepoLocator)
	reqFactory := requirements.NewFactory(deps.termUI, deps.configRepo, deps.apiRepoLocator)
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

func handlePanics() {
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
