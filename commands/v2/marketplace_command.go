package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type MarketplaceCommand struct {
	ServicePlanInfo string      `short:"s" description:"Show plan details for a particular service offering"`
	usage           interface{} `usage:"CF_NAME marketplace [-s SERVICE]"`
}

func (_ MarketplaceCommand) Setup() error {
	return nil
}

func (_ MarketplaceCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
