package resources

type Metadata struct {
	Guid string `json:"guid"`
	Url  string `json:"url,omitempty"`
}

type Resource struct {
	Metadata Metadata
}
