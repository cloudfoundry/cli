package ccv2

type Resource struct {
	Filename string `json:"fn,omitempty"`
	Size     int64  `json:"size"`
	SHA1     string `json:"sha1"`
	Mode     string `json:"mode,omitempty"`
}
