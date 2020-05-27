package constant

// SecurityGroupLifecycle represents the lifecycle phase of a security group
// binding.
type SecurityGroupLifecycle string

const (
	// SecurityGroupLifecycleRunning indicates the lifecycle phase running.
	SecurityGroupLifecycleRunning SecurityGroupLifecycle = "running"

	// SecurityGroupLifecycleStaging indicates the lifecycle phase staging.
	SecurityGroupLifecycleStaging SecurityGroupLifecycle = "staging"
)
