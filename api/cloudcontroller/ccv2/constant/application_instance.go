package constant

// ApplicationInstanceState reflects the state of the individual app
// instance.
type ApplicationInstanceState string

const (
	// ApplicationInstanceCrashed represents an application instance in the
	// crashed state.
	ApplicationInstanceCrashed ApplicationInstanceState = "CRASHED"
	// ApplicationInstanceDown represents an application instance in the down
	// state.
	ApplicationInstanceDown ApplicationInstanceState = "DOWN"
	// ApplicationInstanceFlapping represents an application instance that keeps
	// failing after it starts.
	ApplicationInstanceFlapping ApplicationInstanceState = "FLAPPING"
	// ApplicationInstanceRunning represents an application instance that is
	// currently running.
	ApplicationInstanceRunning ApplicationInstanceState = "RUNNING"
	// ApplicationInstanceStarting represents an application that is the process
	// of starting.
	ApplicationInstanceStarting ApplicationInstanceState = "STARTING"
	// ApplicationInstanceUnknown represents a state that cannot be determined.
	ApplicationInstanceUnknown ApplicationInstanceState = "UNKNOWN"
)
