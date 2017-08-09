package uaa

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/uaa/internal"
)

func (client *Client) GetSSHPasscode(accessToken string, sshOAuthClient string) (string, error) {
	queryValues := url.Values{}
	queryValues.Add("response_type", "code")
	queryValues.Add("client_id", sshOAuthClient)

	request, err := client.newRequest(requestOptions{
		RequestName: internal.GetSSHPasscodeRequest,
		Query:       queryValues,
	})
	if err != nil {
		return "", err
	}

	response := Response{}
	err = client.connection.Make(request, &response)
	if err != nil {
		return "", err
	}

	locationURL, err := response.HTTPResponse.Location()
	if err != nil {
		return "", err
	}

	return locationURL.Query().Get("code"), nil
}
