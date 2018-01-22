package models

type Buildpack struct {
	GUID     string
	Name     string
	Stack    string
	Position *int
	Enabled  *bool
	Key      string
	Filename string
	Locked   *bool
}
