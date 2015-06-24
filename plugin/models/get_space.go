package plugin_models

type GetSpace_Model struct {
	GetSpaces_Model
	Organization     GetOrgs_Model
	Applications     []GetAppsModel
	ServiceInstances []ServiceInstanceSummary
	Domains          []DomainFields
	SecurityGroups   []SecurityGroupFields
	SpaceQuota       SpaceQuotaFields
}
