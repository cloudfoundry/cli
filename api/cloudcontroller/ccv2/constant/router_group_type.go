package constant

// RouterGroupType is an enumeration of all possible router group types.
type RouterGroupType string

const (
	// TCPRouterGroup represents a TCP router group.
	TCPRouterGroup RouterGroupType = "tcp"
	// HTTPRouterGroup represents a HTTP router group.
	HTTPRouterGroup RouterGroupType = "http"
)
