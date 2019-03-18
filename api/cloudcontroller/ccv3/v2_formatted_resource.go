package ccv3

import (
	"encoding/json"
	"os"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// V2FormattedResource represents a Cloud Controller Resource that still has the same shape as the V2 Resource.
// The v3 package upload endpoint understands both the V2 shape and the new V3 shape.
// The v3 resource matching endpoint only understands the new V3 shape.
//
// Deprecated: Use Resource going forward. We anticipate that this struct will only be used
// by the v6 cli's v3-push command, which is experimental.
type V2FormattedResource struct {

	// Filename is the name of the resource.
	Filename string `json:"fn"`

	// Mode is the operating system file mode (aka file permissions) of the
	// resource.
	Mode os.FileMode `json:"mode"`

	// SHA1 represents the SHA-1 hash of the resource.
	SHA1 string `json:"sha1"`

	// Size represents the file size of the resource.
	Size int64 `json:"size"`
}

// MarshalJSON converts a resource into a Cloud Controller V2FormattedResource.
func (r V2FormattedResource) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON helps unmarshal a Cloud Controller V2FormattedResource response.
func (r *V2FormattedResource) UnmarshalJSON(data []byte) error {
	var ccResource struct {
		Filename string `json:"fn,omitempty"`
		Mode     string `json:"mode,omitempty"`
		SHA1     string `json:"sha1"`
		Size     int64  `json:"size"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccResource)
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
