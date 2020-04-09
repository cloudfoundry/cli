package domain

const (
	PermissionRouteForwarding = RequiredPermission("route_forwarding")
	PermissionSyslogDrain     = RequiredPermission("syslog_drain")
	PermissionVolumeMount     = RequiredPermission("volume_mount")

	additionalMetadataName = "AdditionalMetadata"
)

type Service struct {
	ID                   string                  `json:"id"`
	Name                 string                  `json:"name"`
	Description          string                  `json:"description"`
	Bindable             bool                    `json:"bindable"`
	InstancesRetrievable bool                    `json:"instances_retrievable,omitempty"`
	BindingsRetrievable  bool                    `json:"bindings_retrievable,omitempty"`
	Tags                 []string                `json:"tags,omitempty"`
	PlanUpdatable        bool                    `json:"plan_updateable"`
	Plans                []ServicePlan           `json:"plans"`
	Requires             []RequiredPermission    `json:"requires,omitempty"`
	Metadata             *ServiceMetadata        `json:"metadata,omitempty"`
	DashboardClient      *ServiceDashboardClient `json:"dashboard_client,omitempty"`
}

type ServicePlan struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Free            *bool                `json:"free,omitempty"`
	Bindable        *bool                `json:"bindable,omitempty"`
	Metadata        *ServicePlanMetadata `json:"metadata,omitempty"`
	Schemas         *ServiceSchemas      `json:"schemas,omitempty"`
	MaintenanceInfo *MaintenanceInfo     `json:"maintenance_info,omitempty"`
}

type ServiceSchemas struct {
	Instance ServiceInstanceSchema `json:"service_instance,omitempty"`
	Binding  ServiceBindingSchema  `json:"service_binding,omitempty"`
}

type ServiceInstanceSchema struct {
	Create Schema `json:"create,omitempty"`
	Update Schema `json:"update,omitempty"`
}

type ServiceBindingSchema struct {
	Create Schema `json:"create,omitempty"`
}

type Schema struct {
	Parameters map[string]interface{} `json:"parameters"`
}

type RequiredPermission string

func FreeValue(v bool) *bool {
	return &v
}

func BindableValue(v bool) *bool {
	return &v
}

type ServiceDashboardClient struct {
	ID          string `json:"id"`
	Secret      string `json:"secret"`
	RedirectURI string `json:"redirect_uri"`
}
