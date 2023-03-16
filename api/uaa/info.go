package uaa

// Info represents a GET response from a login server
type Info struct {
	Links struct {
		UAA   string `json:"uaa"`
		Login string `json:"login"`
	} `json:"links"`
}

// LoginLink is the URL to the login server.
func (info Info) LoginLink() string {
	return info.Links.Login
}

// UAALink is the URL to the UAA server.
func (info Info) UAALink() string {
	return info.Links.UAA
}

// NewInfo returns back a new
func NewInfo(uaaURL string, loginURL string) Info {
	var info Info
	info.Links.Login = loginURL
	info.Links.UAA = uaaURL
	return info
}
