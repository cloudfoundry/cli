package resources

type LastOperationType string

const (
	CreateOperation LastOperationType = "create"
	UpdateOperation LastOperationType = "update"
	DeleteOperation LastOperationType = "delete"
)

type LastOperationState string

const (
	OperationInProgress LastOperationState = "in progress"
	OperationSucceeded  LastOperationState = "succeeded"
	OperationFailed     LastOperationState = "failed"
)

type LastOperation struct {
	// Type is either "create", "update" or "delete"
	Type LastOperationType `json:"type,omitempty"`
	// State is either "in progress", "succeeded", or "failed"
	State LastOperationState `json:"state,omitempty"`
	// Description contains more details
	Description string `json:"description,omitempty"`
	// CreatedAt is the time when the operation started
	CreatedAt string `json:"created_at,omitempty"`
	// UpdatedAt is the time when the operation was last updated
	UpdatedAt string `json:"updated_at,omitempty"`
}

func (l LastOperation) OmitJSONry() bool {
	return l == LastOperation{}
}
