package v3action

import (
	"strings"
)

// LogCacheURL gets the log-cache URL, either directly from ccClientV3.Info, or by
// getting the target and s/api/logcache/
// This works around the bug logged in story https://www.pivotaltracker.com/story/show/170138644

func (actor Actor) LogCacheURL() string {
	info, _, _, err := actor.CloudControllerClient.GetInfo()
	if err == nil {
		logCacheURL := info.LogCache()
		if logCacheURL != "" {
			return logCacheURL
		}
	}
	return strings.Replace(actor.Config.Target(), "api", "log-cache", 1)
}
