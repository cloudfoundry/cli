package resources

// ServiceAccount represents a space-owned Cloud Controller service account.
type ServiceAccount struct {
	GUID              string        `json:"guid,omitempty"`
	Name              string        `json:"name"`
	Description       string        `json:"description,omitempty"`
	ClientID          string        `json:"client_id,omitempty"`
	CertificateDNSSAN string        `json:"certificate_dns_san,omitempty"`
	Enabled           bool          `json:"enabled,omitempty"`
	Status            string        `json:"status,omitempty"`
	Relationships     Relationships `json:"relationships,omitempty"`
}
