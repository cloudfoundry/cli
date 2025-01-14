package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

// GetApplicationManifest returns a (YAML) manifest for an application and its
// underlying processes.
func (client *Client) GetApplicationManifest(appGUID string) ([]byte, Warnings, error) {
	bytes, warnings, err := client.MakeRequestReceiveRaw(
		ccv3internal.GetApplicationManifestRequest,
		internal.Params{"app_guid": appGUID},
		"application/x-yaml",
	)

	return bytes, warnings, err
}

func (client *Client) GetSpaceManifestDiff(spaceGUID string, rawManifest []byte) (resources.ManifestDiff, Warnings, error) {
	var responseBody resources.ManifestDiff

	_, warnings, err := client.MakeRequestSendRaw(
		ccv3internal.PostSpaceDiffManifestRequest,
		internal.Params{"space_guid": spaceGUID},
		rawManifest,
		"application/x-yaml",
		&responseBody,
	)

	return responseBody, warnings, err
}
