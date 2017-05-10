package ccv2

import "os"

type Resource struct {
	Filename string      `json:"fn,omitempty"`
	Size     int64       `json:"size"`
	SHA1     string      `json:"sha1"`
	Mode     os.FileMode `json:"mode,omitempty"`
}
