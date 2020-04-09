package domain

import (
	"context"
	"encoding/json"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../fakes/auto_fake_service_broker.go -fake-name AutoFakeServiceBroker . ServiceBroker

//Each method of the ServiceBroker interface maps to an individual endpoint of the Open Service Broker API.
//The specification is available here: https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/spec.md
//The OpenAPI documentation is available here: http://petstore.swagger.io/?url=https://raw.githubusercontent.com/openservicebrokerapi/servicebroker/v2.14/openapi.yaml
type ServiceBroker interface {

	// Services gets the catalog of services offered by the service broker
	//   GET /v2/catalog
	Services(ctx context.Context) ([]Service, error)

	// Provision creates a new service instance
	//   PUT /v2/service_instances/{instance_id}
	Provision(ctx context.Context, instanceID string, details ProvisionDetails, asyncAllowed bool) (ProvisionedServiceSpec, error)

	// Deprovision deletes an existing service instance
	//  DELETE /v2/service_instances/{instance_id}
	Deprovision(ctx context.Context, instanceID string, details DeprovisionDetails, asyncAllowed bool) (DeprovisionServiceSpec, error)

	// GetInstance fetches information about a service instance
	//   GET /v2/service_instances/{instance_id}
	GetInstance(ctx context.Context, instanceID string) (GetInstanceDetailsSpec, error)

	// Update modifies an existing service instance
	//  PATCH /v2/service_instances/{instance_id}
	Update(ctx context.Context, instanceID string, details UpdateDetails, asyncAllowed bool) (UpdateServiceSpec, error)

	// LastOperation fetches last operation state for a service instance
	//   GET /v2/service_instances/{instance_id}/last_operation
	LastOperation(ctx context.Context, instanceID string, details PollDetails) (LastOperation, error)

	// Bind creates a new service binding
	//   PUT /v2/service_instances/{instance_id}/service_bindings/{binding_id}
	Bind(ctx context.Context, instanceID, bindingID string, details BindDetails, asyncAllowed bool) (Binding, error)

	// Unbind deletes an existing service binding
	//   DELETE /v2/service_instances/{instance_id}/service_bindings/{binding_id}
	Unbind(ctx context.Context, instanceID, bindingID string, details UnbindDetails, asyncAllowed bool) (UnbindSpec, error)

	// GetBinding fetches an existing service binding
	//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
	GetBinding(ctx context.Context, instanceID, bindingID string) (GetBindingSpec, error)

	// LastBindingOperation fetches last operation state for a service binding
	//   GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation
	LastBindingOperation(ctx context.Context, instanceID, bindingID string, details PollDetails) (LastOperation, error)
}

type LastOperation struct {
	State       LastOperationState `json:"state"`
	Description string             `json:"description"`
}

type LastOperationState string

const (
	InProgress LastOperationState = "in progress"
	Succeeded  LastOperationState = "succeeded"
	Failed     LastOperationState = "failed"
)

type VolumeMount struct {
	Driver       string       `json:"driver"`
	ContainerDir string       `json:"container_dir"`
	Mode         string       `json:"mode"`
	DeviceType   string       `json:"device_type"`
	Device       SharedDevice `json:"device"`
}

type SharedDevice struct {
	VolumeId    string                 `json:"volume_id"`
	MountConfig map[string]interface{} `json:"mount_config"`
}

type ProvisionDetails struct {
	ServiceID        string           `json:"service_id"`
	PlanID           string           `json:"plan_id"`
	OrganizationGUID string           `json:"organization_guid"`
	SpaceGUID        string           `json:"space_guid"`
	RawContext       json.RawMessage  `json:"context,omitempty"`
	RawParameters    json.RawMessage  `json:"parameters,omitempty"`
	MaintenanceInfo  *MaintenanceInfo `json:"maintenance_info,omitempty"`
}

type ProvisionedServiceSpec struct {
	IsAsync       bool
	AlreadyExists bool
	DashboardURL  string
	OperationData string
}

type DeprovisionDetails struct {
	PlanID    string `json:"plan_id"`
	ServiceID string `json:"service_id"`
	Force     bool   `json:"force"`
}

type DeprovisionServiceSpec struct {
	IsAsync       bool
	OperationData string
}

type GetInstanceDetailsSpec struct {
	ServiceID    string      `json:"service_id"`
	PlanID       string      `json:"plan_id"`
	DashboardURL string      `json:"dashboard_url"`
	Parameters   interface{} `json:"parameters"`
}

type UpdateDetails struct {
	ServiceID       string           `json:"service_id"`
	PlanID          string           `json:"plan_id"`
	RawParameters   json.RawMessage  `json:"parameters,omitempty"`
	PreviousValues  PreviousValues   `json:"previous_values"`
	RawContext      json.RawMessage  `json:"context,omitempty"`
	MaintenanceInfo *MaintenanceInfo `json:"maintenance_info,omitempty"`
}

type PreviousValues struct {
	PlanID          string           `json:"plan_id"`
	ServiceID       string           `json:"service_id"`
	OrgID           string           `json:"organization_id"`
	SpaceID         string           `json:"space_id"`
	MaintenanceInfo *MaintenanceInfo `json:"maintenance_info,omitempty"`
}

type UpdateServiceSpec struct {
	IsAsync       bool
	DashboardURL  string
	OperationData string
}

type PollDetails struct {
	ServiceID     string `json:"service_id"`
	PlanID        string `json:"plan_id"`
	OperationData string `json:"operation"`
}

type BindDetails struct {
	AppGUID       string          `json:"app_guid"`
	PlanID        string          `json:"plan_id"`
	ServiceID     string          `json:"service_id"`
	BindResource  *BindResource   `json:"bind_resource,omitempty"`
	RawContext    json.RawMessage `json:"context,omitempty"`
	RawParameters json.RawMessage `json:"parameters,omitempty"`
}

type BindResource struct {
	AppGuid            string `json:"app_guid,omitempty"`
	SpaceGuid          string `json:"space_guid,omitempty"`
	Route              string `json:"route,omitempty"`
	CredentialClientID string `json:"credential_client_id,omitempty"`
	BackupAgent        bool   `json:"backup_agent,omitempty"`
}

type UnbindDetails struct {
	PlanID    string `json:"plan_id"`
	ServiceID string `json:"service_id"`
}

type UnbindSpec struct {
	IsAsync       bool
	OperationData string
}

type Binding struct {
	IsAsync         bool          `json:"is_async"`
	AlreadyExists   bool          `json:"already_exists"`
	OperationData   string        `json:"operation_data"`
	Credentials     interface{}   `json:"credentials"`
	SyslogDrainURL  string        `json:"syslog_drain_url"`
	RouteServiceURL string        `json:"route_service_url"`
	BackupAgentURL  string        `json:"backup_agent_url,omitempty"`
	VolumeMounts    []VolumeMount `json:"volume_mounts"`
}

type GetBindingSpec struct {
	Credentials     interface{}
	SyslogDrainURL  string
	RouteServiceURL string
	VolumeMounts    []VolumeMount
	Parameters      interface{}
}

func (d ProvisionDetails) GetRawContext() json.RawMessage {
	return d.RawContext
}

func (d ProvisionDetails) GetRawParameters() json.RawMessage {
	return d.RawParameters
}

func (d BindDetails) GetRawContext() json.RawMessage {
	return d.RawContext
}

func (d BindDetails) GetRawParameters() json.RawMessage {
	return d.RawParameters
}

func (d UpdateDetails) GetRawContext() json.RawMessage {
	return d.RawContext
}

func (d UpdateDetails) GetRawParameters() json.RawMessage {
	return d.RawParameters
}

type DetailsWithRawParameters interface {
	GetRawParameters() json.RawMessage
}

type DetailsWithRawContext interface {
	GetRawContext() json.RawMessage
}
