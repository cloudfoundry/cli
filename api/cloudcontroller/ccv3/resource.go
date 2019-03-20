package ccv3

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Checksum struct {
	Value string `json:"value"`
}

type Resource struct {
	// FilePath is the path of the resource.
	FilePath string `json:"path"`

	// Mode is the operating system file mode (aka file permissions) of the
	// resource.
	Mode os.FileMode `json:"mode"`

	// SHA1 represents the SHA-1 hash of the resource.
	Checksum Checksum `json:"checksum"`

	// Size represents the file size of the resource.
	SizeInBytes int64 `json:"size_in_bytes"`
}

// MarshalJSON converts a resource into a Cloud Controller Resource.
func (r Resource) MarshalJSON() ([]byte, error) {
	var ccResource struct {
		FilePath    string   `json:"path,omitempty"`
		Mode        string   `json:"mode,omitempty"`
		Checksum    Checksum `json:"checksum"`
		SizeInBytes int64    `json:"size_in_bytes"`
	}

	ccResource.FilePath = r.FilePath
	ccResource.SizeInBytes = r.SizeInBytes
	ccResource.Checksum = r.Checksum
	ccResource.Mode = strconv.FormatUint(uint64(r.Mode), 10)
	return json.Marshal(ccResource)
}

func (r Resource) ToV2FormattedResource() V2FormattedResource {
	return V2FormattedResource{
		Filename: r.FilePath,
		Mode:     r.Mode,
		SHA1:     r.Checksum.Value,
		Size:     r.SizeInBytes,
	}
}

// UnmarshalJSON helps unmarshal a Cloud Controller Resource response.
func (r *Resource) UnmarshalJSON(data []byte) error {
	var ccResource struct {
		FilePath    string   `json:"path,omitempty"`
		Mode        string   `json:"mode,omitempty"`
		Checksum    Checksum `json:"checksum"`
		SizeInBytes int64    `json:"size_in_bytes"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccResource)
	if err != nil {
		return err
	}

	r.FilePath = ccResource.FilePath
	r.SizeInBytes = ccResource.SizeInBytes
	r.Checksum = ccResource.Checksum
	mode, err := strconv.ParseUint(ccResource.Mode, 10, 32)
	if err != nil {
		return err
	}

	r.Mode = os.FileMode(mode)
	return nil
}

func (client Client) ResourceMatch(resources []Resource) ([]Resource, Warnings, error) {
	bodyBytes, err := json.Marshal(map[string][]Resource{"resources": resources})
	if err != nil {
		return []Resource{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostResourceMatchesRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return []Resource{}, nil, err
	}

	var matchedResources map[string][]Resource
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &matchedResources,
	}
	err = client.connection.Make(request, &response)

	return matchedResources["resources"], response.Warnings, err
}
