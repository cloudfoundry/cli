package ccv2

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type Resource struct {
	Filename string      `json:"fn"`
	Mode     os.FileMode `json:"mode"`
	SHA1     string      `json:"sha1"`
	Size     int64       `json:"size"`
}

func (r *Resource) UnmarshalJSON(rawJSON []byte) error {
	var ccResource struct {
		Filename string `json:"fn,omitempty"`
		Mode     string `json:"mode,omitempty"`
		SHA1     string `json:"sha1"`
		Size     int64  `json:"size"`
	}

	err := json.Unmarshal(rawJSON, &ccResource)
	if err != nil {
		return err
	}

	r.Filename = ccResource.Filename
	r.Size = ccResource.Size
	r.SHA1 = ccResource.SHA1
	mode, err := strconv.ParseUint(ccResource.Mode, 8, 32)
	if err != nil {
		return err
	}

	r.Mode = os.FileMode(mode)
	return nil
}

func (r Resource) MarshalJSON() ([]byte, error) {
	var ccResource struct {
		Filename string `json:"fn,omitempty"`
		Mode     string `json:"mode,omitempty"`
		SHA1     string `json:"sha1"`
		Size     int64  `json:"size"`
	}

	ccResource.Filename = r.Filename
	ccResource.Size = r.Size
	ccResource.SHA1 = r.SHA1
	ccResource.Mode = strconv.FormatUint(uint64(r.Mode), 8)
	return json.Marshal(ccResource)
}

// ResourceMatch returns the resources that exist on the cloud foundry instance
// from the set of resources given.
func (client *Client) ResourceMatch(resourcesToMatch []Resource) ([]Resource, Warnings, error) {
	body, err := json.Marshal(resourcesToMatch)
	if err != nil {
		return nil, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutResourceMatch,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return nil, nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	var matchedResources []Resource
	response := cloudcontroller.Response{
		Result: &matchedResources,
	}

	err = client.connection.Make(request, &response)
	return matchedResources, response.Warnings, err
}
