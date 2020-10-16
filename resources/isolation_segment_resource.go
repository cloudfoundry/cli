package resources

// IsolationSegment represents a Cloud Controller Isolation Segment.
type IsolationSegment struct {
	//GUID is the unique ID of the isolation segment.
	GUID string `json:"guid,omitempty"`
	//Name is the name of the isolation segment.
	Name string `json:"name"`
}
