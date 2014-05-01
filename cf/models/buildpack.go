package models

type Buildpack struct {
	Guid     string
	Name     string
	Position *int
	Enabled  *bool
	Key      string
	Filename string
	Locked   *bool
}
