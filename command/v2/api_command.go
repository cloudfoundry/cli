package v2

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/configaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . APIConfigActor

type APIConfigActor interface {
	ClearTarget()
	SetTarget(settings configaction.TargetSettings) (configaction.Warnings, error)
}

type ApiCommand struct {
	OptionalArgs      flag.APITarget `positional-args:"yes"`
	SkipSSLValidation bool           `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	Unset             bool           `long:"unset" description:"Remove all api endpoint targeting"`
	usage             interface{}    `usage:"CF_NAME api [URL]"`
	relatedCommands   interface{}    `related_commands:"auth, login, target"`

	UI     command.UI
	Actor  APIConfigActor
	Config command.Config
}

func (cmd *ApiCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Actor = configaction.NewActor(config, ccv2.NewClient(config.BinaryName(), config.BinaryVersion()))
	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd *ApiCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTable("", [][]string{
		{cmd.UI.TranslateText("API endpoint:"), cmd.Config.Target()},
		{cmd.UI.TranslateText("API version:"), cmd.Config.APIVersion()},
	}, 3)

	user, err := cmd.Config.CurrentUser()
	if user.Name == "" {
		cmd.UI.DisplayText("Not logged in. Use '{{.CFLoginCommand}}' to log in.", map[string]interface{}{
			"CFLoginCommand": fmt.Sprintf("%s login", cmd.Config.BinaryName()),
		})
	}
	return err
}

func (cmd *ApiCommand) ClearTarget() error {
	cmd.UI.DisplayTextWithFlavor("Unsetting api endpoint...")
	cmd.Actor.ClearTarget()
	cmd.UI.DisplayOK()
	return nil
}

func (cmd *ApiCommand) setAPI() error {
	cmd.UI.DisplayTextWithFlavor("Setting api endpoint to {{.Endpoint}}...", map[string]interface{}{
		"Endpoint": cmd.OptionalArgs.URL,
	})

	apiURL := processURL(cmd.OptionalArgs.URL)

	_, err := cmd.Actor.SetTarget(configaction.TargetSettings{
		URL:               apiURL,
		SkipSSLValidation: cmd.SkipSSLValidation,
		DialTimeout:       cmd.Config.DialTimeout(),
	})
	if err != nil {
		return shared.HandleError(err)
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
