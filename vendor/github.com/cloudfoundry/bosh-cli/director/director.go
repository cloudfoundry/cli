package director

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DirectorImpl struct {
	client Client
}

func (d DirectorImpl) WithContext(id string) Director {
	return DirectorImpl{client: d.client.WithContext(id)}
}

func (d DirectorImpl) EnableResurrection(enabled bool) error {
	return d.client.EnableResurrectionAll(enabled)
}

func (d DirectorImpl) CleanUp(all bool) error {
	return d.client.CleanUp(all)
}

func (d DirectorImpl) DownloadResourceUnchecked(blobstoreID string, out io.Writer) error {
	return d.client.DownloadResourceUnchecked(blobstoreID, out)
}

func (c Client) EnableResurrectionAll(enabled bool) error {
	body := map[string]bool{"resurrection_paused": !enabled}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	_, _, err = c.clientRequest.RawPut("/resurrection", reqBody, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Changing VM resurrection state for all")
	}

	return nil
}

func (c Client) CleanUp(all bool) error {
	body := map[string]interface{}{
		"config": map[string]bool{"remove_all": all},
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	_, err = c.taskClientRequest.PostResult("/cleanup", reqBody, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Cleaning up resources")
	}

	return nil
}

func (c Client) DownloadResourceUnchecked(blobstoreID string, out io.Writer) error {
	path := fmt.Sprintf("/resources/%s", blobstoreID)

	_, _, err := c.clientRequest.RawGet(path, out, nil)
	if err != nil {
		return bosherr.WrapErrorf(err, "Downloading resource '%s'", blobstoreID)
	}

	return nil
}
