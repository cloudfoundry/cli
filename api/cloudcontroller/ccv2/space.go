package ccv2

// Space represents a Cloud Controller Space.
type Space struct {
	GUID     string
	Name     string
	AllowSSH bool
}

func (client *Client) GetSpaces(queries []Query) ([]Space, Warnings, error) {
	return nil, nil, nil
}
