package clissh

type TTYRequest int

const (
	RequestTTYAuto TTYRequest = iota
	RequestTTYNo
	RequestTTYYes
	RequestTTYForce
)
