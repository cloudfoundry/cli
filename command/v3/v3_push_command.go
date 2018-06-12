package v3

import (
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

type V3PushCommand struct {
	RequiredArgs   flag.AppName                `positional-args:"yes"`
	Buildpacks     []string                    `short:"b" description:"Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	DockerImage    flag.DockerImage            `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	DockerUsername string                      `long:"docker-username" description:"Repository username; used with password from environment variable CF_DOCKER_PASSWORD"`
	NoRoute        bool                        `long:"no-route" description:"Do not map a route to this app"`
	NoStart        bool                        `long:"no-start" description:"Do not stage and start the app after pushing"`
	AppPath        flag.PathWithExistenceCheck `short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	dockerPassword interface{}                 `environmentName:"CF_DOCKER_PASSWORD" environmentDescription:"Password used for private docker repository"`

	usage               interface{} `usage:"cf v3-push APP_NAME [-b BUILDPACK]... [-p APP_PATH] [--no-route] [--no-start]\n   cf v3-push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME] [--no-route] [--no-start]"`
	envCFStagingTimeout interface{} `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{} `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI                  command.UI
	Config              command.Config
	NOAAClient          v3action.NOAAClient
	SharedActor         command.SharedActor
	AppSummaryDisplayer shared.AppSummaryDisplayer
	PackageDisplayer    shared.PackageDisplayer

	OriginalActor       OriginalV3PushActor
	OriginalV2PushActor OriginalV2PushActor
}

func (cmd *V3PushCommand) Setup(config command.Config, ui command.UI) error {
	if !config.Experimental() {
		return cmd.OriginalSetup(config, ui)
	}
	return nil
}

func (cmd V3PushCommand) Execute(args []string) error {
	if !cmd.Config.Experimental() {
		return cmd.OriginalExecute(args)
	}
	return nil
}
