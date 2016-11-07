package v2

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/actors/configactions"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
	"code.cloudfoundry.org/cli/commands/v2/common"
)

//go:generate counterfeiter . APIConfigActor

type APIConfigActor interface {
	ClearTarget()
	SetTarget(settings configactions.TargetSettings) (configactions.Warnings, error)
}

type ApiCommand struct {
	OptionalArgs      flags.APITarget `positional-args:"yes"`
	SkipSSLValidation bool            `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	Unset             bool            `long:"unset" description:"Remove all api endpoint targeting"`
	usage             interface{}     `usage:"CF_NAME api [URL]"`
	relatedCommands   interface{}     `related_commands:"auth, login, target"`

	UI     commands.UI
	Actor  APIConfigActor
	Config commands.Config
}

func (cmd *ApiCommand) Setup(config commands.Config, ui commands.UI) error {
	cmd.Actor = configactions.NewActor(config, common.NewCloudControllerClient(config.BinaryName()))
	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd *ApiCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}

	cmd.UI.DisplayText(ExperimentalWarning)
	cmd.UI.DisplayNewline()

	if cmd.Unset {
		return cmd.ClearTarget()
	}

	if cmd.OptionalArgs.URL != "" {
		err := cmd.setAPI()
		if err != nil {
			return err
		}
	}

	if cmd.Config.Target() == "" {
		cmd.UI.DisplayText("No api endpoint set. Use '{{.Name}}' to set an endpoint", map[string]interface{}{
			"Name": "cf api",
		})
		return nil
	}

	return DisplayCurrentTargetInformation(cmd.Config, cmd.UI)
}

func (cmd *ApiCommand) ClearTarget() error {
	cmd.UI.DisplayHeaderFlavorText("Unsetting api endpoint...")
	cmd.Actor.ClearTarget()
	cmd.UI.DisplayOK()
	return nil
}

func DisplayCurrentTargetInformation(config commands.Config, commandUI commands.UI) error {
	user, err := config.CurrentUser()
	if err != nil {
		return err
	}

	commandUI.DisplayPair("API endpoint", config.Target())
	commandUI.DisplayPair("API version", config.APIVersion())
	commandUI.DisplayPair("User", user.Name)
	commandUI.DisplayPair("Org", config.TargetedOrganization().Name)
	commandUI.DisplayPair("Space", config.TargetedSpace().Name)
	return nil
}

func (cmd *ApiCommand) setAPI() error {
	cmd.UI.DisplayHeaderFlavorText("Setting api endpoint to {{.Endpoint}}...", map[string]interface{}{
		"Endpoint": cmd.OptionalArgs.URL,
	})

	apiURL := processURL(cmd.OptionalArgs.URL)

	_, err := cmd.Actor.SetTarget(configactions.TargetSettings{
		URL:               apiURL,
		SkipSSLValidation: cmd.SkipSSLValidation,
		DialTimeout:       cmd.Config.DialTimeout(),
	})
	if err != nil {
		return common.HandleError(err)
	}

	if strings.HasPrefix(apiURL, "http:") {
		cmd.UI.DisplayText("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended")
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	return nil
}

func processURL(apiURL string) string {
	if !strings.HasPrefix(apiURL, "http") {
		return fmt.Sprintf("https://%s", apiURL)

	}
	return apiURL
}
