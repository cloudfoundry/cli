package cf

const (
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
