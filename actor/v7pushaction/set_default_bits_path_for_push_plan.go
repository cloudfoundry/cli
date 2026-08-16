package v7pushaction

import (
	"log/slog"
	"os"
)

func SetDefaultBitsPathForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	if pushPlan.BitsPath == "" && pushPlan.DropletPath == "" && pushPlan.DockerImageCredentials.Path == "" {
		var err error
		pushPlan.BitsPath, err = os.Getwd()
		slog.Debug("using current directory for bits path", "path", pushPlan.BitsPath)
		if err != nil {
			return pushPlan, err
		}
	}
	return pushPlan, nil
}
