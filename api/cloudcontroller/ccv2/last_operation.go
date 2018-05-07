package ccv2

// LastOperation is the status of the last operation requested on a service
// instance.
type LastOperation struct {
	// Type is the type of operation that was last performed or currently being
	// performed on the service instance.
	Type string

	// State is the status of the last operation or current operation being
	// performed on the service instance.
	State string

	// Description is the service broker-provided description of the operation.
	Description string

	// UpdatedAt is the timestamp that the Cloud Controller last checked the
	// service instance state from the broker.
	UpdatedAt string

	// CreatedAt is the timestamp that the Cloud Controller created the service
	// instance from the broker.
	CreatedAt string
}
