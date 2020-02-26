package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// GetApplicationManifest returns a (YAML) manifest for an application and its
// underlying processes.
func (client *Client) GetApplicationManifest(appGUID string) ([]byte, Warnings, error) {
	bytes, warnings, err := client.MakeRequestReceiveRaw(
		internal.GetApplicationManifestRequest,
		internal.Params{"app_guid": appGUID},
		"application/x-yaml",
	)

	return bytes, warnings, err
}
