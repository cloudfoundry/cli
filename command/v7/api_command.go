package v7

import (
	"code.cloudfoundry.org/clock"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . APIActor

type APIActor interface {
	ClearTarget()
	SetTarget(settings v7action.TargetSettings) (v7action.Warnings, error)
}

type APICommand struct {
	OptionalArgs  flag.APITarget `positional-args:"yes"`
	SkipSSLValidation bool           `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	Unset             bool           `long:"unset" description:"Remove all api endpoint targeting"`
	usage             interface{}    `usage:"CF_NAME api [URL]"`
	relatedCommands   interface{}    `related_commands:"auth, login, target"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       APIActor
}

func (cmd *APICommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config


	ccClient, _ := shared.NewWrappedCloudControllerClient(config, ui)
	cmd.Actor = v7action.NewActor(ccClient, config, nil,nil, clock.NewClock())
	return nil
}

func (cmd *APICommand) Execute(args []string) error {
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

	if cmd.Config.APIVersion() != "" {
		err := command.WarnIfAPIVersionBelowSupportedMinimum(cmd.Config.APIVersion(), cmd.UI)
		if err != nil {
			return err
		}
	}

	cmd.UI.DisplayKeyValueTable("", [][]string{
		{cmd.UI.TranslateText("api endpoint:"), cmd.Config.Target()},
		{cmd.UI.TranslateText("api version:"), cmd.Config.APIVersion()},
	}, 3)

	user, err := cmd.Config.CurrentUser()
	if user.Name == "" {
		command.DisplayNotLoggedInText(cmd.Config.BinaryName(), cmd.UI)
	}
	return err
}

func (cmd *APICommand) ClearTarget() error {
	cmd.UI.DisplayTextWithFlavor("Unsetting api endpoint...")
	cmd.Actor.ClearTarget()
	cmd.UI.DisplayOK()
	return nil
}

func (cmd *APICommand) setAPI() error {
	cmd.UI.DisplayTextWithFlavor("Setting api endpoint to {{.Endpoint}}...", map[string]interface{}{
		"Endpoint": cmd.OptionalArgs.URL,
	})

	apiURL := processURL(cmd.OptionalArgs.URL)

	_, err := cmd.Actor.SetTarget(v7action.TargetSettings{
		URL:               apiURL,
		SkipSSLValidation: cmd.SkipSSLValidation,
		DialTimeout:       cmd.Config.DialTimeout(),
	})
	if err != nil {
		
		return err
	}

	if strings.HasPrefix(apiURL, "http:") {
		panic("good")
		cmd.UI.DisplayText("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended")
	}

	cmd.UI.DisplayOK()
	return nil
}

func processURL(apiURL string) string {
	if !strings.HasPrefix(apiURL, "http") {
		return fmt.Sprintf("https://%s", apiURL)

	}
	return apiURL
}
