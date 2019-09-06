package v7pushaction

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func SetDefaultBitsPathForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	if pushPlan.BitsPath == "" && pushPlan.DropletPath == "" && pushPlan.DockerImageCredentials.Path == "" {
		var err error
		pushPlan.BitsPath, err = os.Getwd()
		log.WithField("path", pushPlan.BitsPath).Debug("using current directory for bits path")
		if err != nil {
			return pushPlan, err
		}
	}
	return pushPlan, nil
}
