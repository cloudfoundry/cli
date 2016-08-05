package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type LoginCommand struct {
	APIEndpoint       string `short:"a" description:"API endpoint (e.g. https://api.example.com)"`
	Username          string `short:"u" description:"Username"`
	Password          string `short:"p" description:"Password"`
	Organization      string `short:"o" description:"Org"`
	Space             string `short:"s" description:"Space"`
	SSO               bool   `long:"sso" description:"Use a one-time password to login"`
	SkipSSLValidation bool   `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
}

func (_ LoginCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
