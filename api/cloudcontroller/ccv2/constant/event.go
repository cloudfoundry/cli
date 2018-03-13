package constant

// EventType is the type of event that occurred in cloud foundry.
type EventType string

const (
	EventTypeAuditApplicationCreate                 EventType = "audit.app.create"
	EventTypeAuditApplicationDelete                 EventType = "audit.app.delete-request"
	EventTypeAuditApplicationMapRoute               EventType = "audit.app.map-route"
	EventTypeAuditApplicationProcessDelete          EventType = "audit.app.process.delete"
	EventTypeAuditApplicationUnmapRoute             EventType = "audit.app.unmap-route"
	EventTypeAuditApplicationUpdate                 EventType = "audit.app.update"
	EventTypeAuditRouteCreate                       EventType = "audit.route.create"
	EventTypeAuditRouteDelete                       EventType = "audit.route.delete-request"
	EventTypeAuditRouteUpdate                       EventType = "audit.route.update"
	EventTypeAuditServiceBindingCreate              EventType = "audit.service_binding.create"
	EventTypeAuditServiceBindingDelete              EventType = "audit.service_binding.delete"
	EventTypeAuditServiceInstanceCreate             EventType = "audit.service_instance.create"
	EventTypeAuditServiceInstanceDelete             EventType = "audit.service_instance.delete"
	EventTypeAuditServiceInstanceUpdate             EventType = "audit.service_instance.update"
	EventTypeAuditSpaceCreate                       EventType = "audit.space.create"
	EventTypeAuditSpaceDelete                       EventType = "audit.space.delete-request"
	EventTypeAuditSpaceUpdate                       EventType = "audit.space.update"
	EventTypeAuditUserProvidedServiceInstanceCreate EventType = "audit.user_provided_service_instance.create"
	EventTypeAuditUserProvidedServiceInstanceDelete EventType = "audit.user_provided_service_instance.delete"
	EventTypeAuditUserProvidedServiceInstanceUpdate EventType = "audit.user_provided_service_instance.update"
)
