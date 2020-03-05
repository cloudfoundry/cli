package ccv3

type IncludedResources struct {
	Users         []User         `json:"users,omitempty"`
	Organizations []Organization `json:"organizations,omitempty"`
	Spaces        []Space        `json:"spaces,omitempty"`
}
