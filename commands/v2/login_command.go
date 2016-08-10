package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type LoginCommand struct {
	APIEndpoint       string      `short:"a" description:"API endpoint (e.g. https://api.example.com)"`
	Username          string      `short:"u" description:"Username"`
	Password          string      `short:"p" description:"Password"`
	Organization      string      `short:"o" description:"Org"`
	Space             string      `short:"s" description:"Space"`
	SSO               bool        `long:"sso" description:"Use a one-time password to login"`
	SkipSSLValidation bool        `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	usage             interface{} `usage:"CF_NAME login [-a API_URL] [-u USERNAME] [-p PASSWORD] [-o ORG] [-s SPACE]\n\nWARNING:\n    Providing your password as a command line option is highly discouraged\n    Your password may be visible to others and may be recorded in your shell history\n\nEXAMPLES:\n    CF_NAME login (omit username and password to login interactively -- CF_NAME will prompt for both)\n    CF_NAME login -u name@example.com -p pa55woRD (specify username and password as arguments)\n    CF_NAME login -u name@example.com -p \"my password\" (use quotes for passwords with a space)\n    CF_NAME login -u name@example.com -p \"\\\"password\\\"\" (escape quotes if used in password)\n    CF_NAME login --sso (CF_NAME will provide a url to obtain a one-time password to login)"`
}

func (_ LoginCommand) Setup() error {
	return nil
}

func (_ LoginCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
