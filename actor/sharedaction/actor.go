// Package sharedaction handles all operations that do not require a cloud
// controller
package sharedaction

// Actor handles all shared actions
type Actor struct{}

// NewActor returns an Actor with default settings
func NewActor() Actor {
	return Actor{}
}
