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
	Size     int64       `json:"size"`
	SHA1     string      `json:"sha1"`
	Mode     os.FileMode `json:"mode"`
}

func (r Resource) MarshalJSON() ([]byte, error) {
	var ccResource struct {
		Filename string `json:"fn,omitempty"`
		Size     int64  `json:"size"`
		SHA1     string `json:"sha1"`
		Mode     string `json:"mode,omitempty"`
	}

	ccResource.Filename = r.Filename
	ccResource.Size = r.Size
	ccResource.SHA1 = r.SHA1
	ccResource.Mode = strconv.FormatUint(uint64(r.Mode), 8)
	return json.Marshal(ccResource)
}

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

	var matchedResources []Resource
	response := cloudcontroller.Response{
		Result: &matchedResources,
	}

	err = client.connection.Make(request, &response)
	return matchedResources, response.Warnings, err
}
