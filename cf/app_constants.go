package cf

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/i18n"
)

const (
	Version     = "BUILT_FROM_SOURCE"
	BuiltOnDate = "BUILT_AT_UNKNOWN_TIME"
)

var (
	t     = i18n.Init("cf", i18n.GetResourcesPath())
	Usage = t("A command line tool to interact with Cloud Foundry")
)

func Name() string {
	return filepath.Base(os.Args[0])
}
