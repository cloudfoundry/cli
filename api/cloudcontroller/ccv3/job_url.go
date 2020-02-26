package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// JobURL is the URL to a given Job.
type JobURL string

// DeleteApplication deletes the app with the given app GUID. Returns back a
// resulting job URL to poll.
func (client *Client) DeleteApplication(appGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteApplicationRequest,
		URIParams:   internal.Params{"app_guid": appGUID},
	})

	return jobURL, warnings, err
}

// UpdateApplicationApplyManifest applies the manifest to the given
// application. Returns back a resulting job URL to poll.
func (client *Client) UpdateApplicationApplyManifest(appGUID string, rawManifest []byte) (JobURL, Warnings, error) {
	responseLocation, warnings, err := client.MakeRequestSendRaw(
		internal.PostApplicationActionApplyManifest,
		internal.Params{"app_guid": appGUID},
		rawManifest,
		"application/x-yaml",
		nil,
	)

	return JobURL(responseLocation), warnings, err
}

// UpdateSpaceApplyManifest - Is there a better name for this, since ...
// -- The Space resource is not actually updated.
// -- Instead what this ApplyManifest may do is to Create or Update Applications instead.

// Applies the manifest to the given space. Returns back a resulting job URL to poll.

// For each app specified in the manifest, the server-side handles:
// (1) Finding or creating this app.
// (2) Applying manifest properties to this app.

func (client *Client) UpdateSpaceApplyManifest(spaceGUID string, rawManifest []byte) (JobURL, Warnings, error) {
	responseLocation, warnings, err := client.MakeRequestSendRaw(
		internal.PostSpaceActionApplyManifestRequest,
		internal.Params{"space_guid": spaceGUID},
		rawManifest,
		"application/x-yaml",
		nil,
	)

	return JobURL(responseLocation), warnings, err
}
