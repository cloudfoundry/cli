package helpers

import (
	"fmt"
)

func GetPackageFirstDroplet(packageGUID string) string {
	session := CF("curl", fmt.Sprintf("v3/packages/%s/droplets", packageGUID))
	bytes := session.Wait("15s").Out.Contents()
	return getGUID(bytes)
}

func GetAppDroplet(appGUID string) string {
	session := CF("curl", fmt.Sprintf("v3/apps/%s/droplets?order_by=-created_at", appGUID))
	bytes := session.Wait("15s").Out.Contents()
	return getGUID(bytes)
}
