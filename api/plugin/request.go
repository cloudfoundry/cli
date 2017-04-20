package plugin

import "net/http"

// newHTTPRequest returns a constructed HTTP.Request with some defaults.
// Defaults are applied when Request options are not filled in.
func (client *Client) newHTTPGetRequest(url string) (*http.Request, error) {
	var request *http.Request
	var err error

	request, err = http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	request.Header = http.Header{}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", client.userAgent)

	return request, nil
}
