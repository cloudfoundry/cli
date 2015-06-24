package plugin_models

type Space struct {
	SpaceSummary
	Organization     GetOrgs_Model
	Applications     []GetAppsModel
	ServiceInstances []ServiceInstanceSummary
	Domains          []DomainFields
	SecurityGroups   []SecurityGroupFields
	SpaceQuota       SpaceQuotaFields
}
