package ccv2

// SecurityGroupRule represents a Cloud Controller Security Group Role.
type SecurityGroupRule struct {
	// Description is a short message discribing the rule.
	Description string

	// Destination is the destination CIDR or range of IPs.
	Destination string

	// Ports is the port or port range.
	Ports string

	// Protocol can be tcp, icmp, udp, all.
	Protocol string
}
