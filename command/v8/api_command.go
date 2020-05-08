package v8

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/clock"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type APICommand struct {
	BaseCommand

	OptionalArgs      flag.APITarget `positional-args:"yes"`
	SkipSSLValidation bool           `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	Unset             bool           `long:"unset" description:"Remove all api endpoint targeting"`
	usage             interface{}    `usage:"CF_NAME api [URL]"`
	relatedCommands   interface{}    `related_commands:"auth, login, target"`
}

func (cmd *APICommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config

	ccClient, _ := shared.NewWrappedCloudControllerClient(config, ui)
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())
	return nil
}

func (cmd *APICommand) Execute(args []string) error {
	if cmd.Unset {
		return cmd.clearTarget()
	}

	if cmd.OptionalArgs.URL != "" {
		return cmd.setAPI()
	}

	return cmd.viewTarget()
}

func (cmd *APICommand) clearTarget() error {
	cmd.UI.DisplayTextWithFlavor("V8: Unsetting API endpoint...")
	cmd.Actor.ClearTarget()
	cmd.UI.DisplayOK()
	return nil
}

func (cmd *APICommand) setAPI() error {
	cmd.UI.DisplayTextWithFlavor("Setting API endpoint to {{.Endpoint}}...", map[string]interface{}{
		"Endpoint": cmd.OptionalArgs.URL,
	})

	apiURL := cmd.processURL(cmd.OptionalArgs.URL)

	_, err := cmd.Actor.SetTarget(v7action.TargetSettings{
		URL:               apiURL,
		SkipSSLValidation: cmd.SkipSSLValidation,
		DialTimeout:       cmd.Config.DialTimeout(),
	})
	if err != nil {
		return err
	}

	if strings.HasPrefix(apiURL, "http:") {
		cmd.UI.DisplayText("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended")
	}

	cmd.UI.DisplayOK()
	return cmd.displayTarget()
}

func (cmd *APICommand) processURL(apiURL string) string {
	if !strings.HasPrefix(apiURL, "http") {
		return fmt.Sprintf("https://%s", apiURL)
	}
	return apiURL
}

func (cmd *APICommand) viewTarget() error {
	if cmd.Config.Target() == "" {
		cmd.UI.DisplayText("No API endpoint set. Use '{{.Name}}' to set an endpoint", map[string]interface{}{
			"Name": "cf api",
		})
		return nil
	}

	return cmd.displayTarget()
}

func (cmd *APICommand) displayTarget() error {
	cmd.UI.DisplayKeyValueTable("", [][]string{
		{cmd.UI.TranslateText("API endpoint:"), cmd.Config.Target()},
		{cmd.UI.TranslateText("API version:"), cmd.Config.APIVersion()},
	}, 3)

	user, err := cmd.Config.CurrentUser()
	if user.Name == "" {
		cmd.UI.DisplayNewline()
		command.DisplayNotLoggedInText(cmd.Config.BinaryName(), cmd.UI)
	}
	return err
}
