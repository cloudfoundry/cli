package v2

var Commands commands

type commands struct {
	AppCommand AppCommand `command:"app"`
}
