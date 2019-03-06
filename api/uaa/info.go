package uaa

// Info represents a GET response from a login server
type Info struct {
	App struct {
		Version string `json:"version"`
	} `json:"app"`
	Links struct {
		UAA   string `json:"uaa"`
		Login string `json:"login"`
	} `json:"links"`
	Prompts map[string][]string `json:"prompts"`
}

// APIVersion is the version of the server.
func (info Info) APIVersion() string {
	return info.App.Version
}

// LoginLink is the URL to the login server.
func (info Info) LoginLink() string {
	return info.Links.Login
}

func (info Info) LoginPrompts() map[string][]string {
	return info.Prompts
}

// UAALink is the URL to the UAA server.
func (info Info) UAALink() string {
	return info.Links.UAA
}

// NewInfo returns back a new
func NewInfo(link string) Info {
	var info Info
	info.Links.Login = link
	info.Links.UAA = link
	return info
}
