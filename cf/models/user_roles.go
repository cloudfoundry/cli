package models

const (
	ORG_USER        = "OrgUser"
	ORG_MANAGER     = "OrgManager"
	BILLING_MANAGER = "BillingManager"
	ORG_AUDITOR     = "OrgAuditor"
	SPACE_MANAGER   = "SpaceManager"
	SPACE_DEVELOPER = "SpaceDeveloper"
	SPACE_AUDITOR   = "SpaceAuditor"
)

var UserInputToOrgRole = map[string]string{
	"OrgManager":     ORG_MANAGER,
	"BillingManager": BILLING_MANAGER,
	"OrgAuditor":     ORG_AUDITOR,
}

var UserInputToSpaceRole = map[string]string{
	"SpaceManager":   SPACE_MANAGER,
	"SpaceDeveloper": SPACE_DEVELOPER,
	"SpaceAuditor":   SPACE_AUDITOR,
}

var SpaceRoleToUserInput = map[string]string{
	SPACE_MANAGER:   "SpaceManager",
	SPACE_DEVELOPER: "SpaceDeveloper",
	SPACE_AUDITOR:   "SpaceAuditor",
}

var QueryParmToOrgRole = map[string]string{
	"manager_guid":         ORG_MANAGER,
	"billing_manager_guid": BILLING_MANAGER,
	"auditor_guid":         ORG_AUDITOR,
}

var QueryParmToSpaceRole = map[string]string{
	"managed_spaces": SPACE_MANAGER,
	"developer_guid": SPACE_DEVELOPER,
	"audited_spaces": SPACE_AUDITOR,
}
