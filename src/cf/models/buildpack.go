package models

type Buildpack struct {
	BasicFields
	Position *int
	Enabled  *bool
	Key      string
	Filename string
	Locked   *bool
}
