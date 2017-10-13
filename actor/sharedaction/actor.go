// Package sharedaction handles all operations that do not require a cloud
// controller
package sharedaction

// Actor handles all shared actions
type Actor struct {
	Config            Config
	SecureShellClient SecureShellClient
}

// NewActor returns an Actor with default settings
func NewActor(config Config, sshClient SecureShellClient) *Actor {
	return &Actor{
		Config:            config,
		SecureShellClient: sshClient,
	}
}
