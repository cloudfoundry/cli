package constant

// ProcessInstanceState is the state of a given process.
type ProcessInstanceState string

const (
	// ProcessInstanceRunning is when the process is running normally.
	ProcessInstanceRunning ProcessInstanceState = "RUNNING"
	// ProcessInstanceCrashed is when the process has crashed.
	ProcessInstanceCrashed ProcessInstanceState = "CRASHED"
	// ProcessInstanceStarting is when the process is starting up.
	ProcessInstanceStarting ProcessInstanceState = "STARTING"
	// ProcessInstanceDown is when the process has gone down.
	ProcessInstanceDown ProcessInstanceState = "DOWN"
)

// ProcessTypeWeb represents the "web" process type.
const ProcessTypeWeb = "web"
