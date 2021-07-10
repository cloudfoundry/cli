package constant

// RoleType is the type of a CCV3 role resource.
type RoleType string

const (
	OrgUserRole           RoleType = "organization_user"
	OrgAuditorRole        RoleType = "organization_auditor"
	OrgManagerRole        RoleType = "organization_manager"
	OrgBillingManagerRole RoleType = "organization_billing_manager"
	SpaceDeveloperRole    RoleType = "space_developer"
	SpaceAuditorRole      RoleType = "space_auditor"
	SpaceManagerRole      RoleType = "space_manager"
	SpaceSupporterRole    RoleType = "space_supporter"
)
