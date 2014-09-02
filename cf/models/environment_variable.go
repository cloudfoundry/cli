package models

func NewEnvironmentVariable(name string, value string) (e EnvironmentVariable) {
	e.Name = name
	e.Value = value
	return
}

type EnvironmentVariable struct {
	Name  string
	Value string
}
