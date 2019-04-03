package v7pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/util/manifestparser"
	log "github.com/sirupsen/logrus"
)

func SetupBitsPathForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	log.Info("determine bits path")
	switch {
	case overrides.ProvidedAppPath != "":
		log.WithField("path", overrides.ProvidedAppPath).Debug("using flag override path for bits path")
		pushPlan.BitsPath = overrides.ProvidedAppPath
	case manifestApp.Path != "":
		log.WithField("path", manifestApp.Path).Debug("using manifest path for bits path")
		pushPlan.BitsPath = manifestApp.Path
	default:
		var err error
		pushPlan.BitsPath, err = os.Getwd()
		log.WithField("path", pushPlan.BitsPath).Debug("using current directory for bits path")
		if err != nil {
			return pushPlan, err
		}
	}
	return pushPlan, nil
}
