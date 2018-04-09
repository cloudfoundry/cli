package constant

// EventType is the type of event that occurred in cloud foundry.
type EventType string

const (
	// EventTypeApplicationCrash denotes an event where an application crashed.
	EventTypeApplicationCrash EventType = "app.crash"

	// EventTypeAuditApplicationCopyBits denotes an event where the CC copies bits
	// from one application to another
	EventTypeAuditApplicationCopyBits EventType = "audit.app.copy-bits"

	// EventTypeAuditApplicationCreate denotes an event where the CC creates an
	// applicaion.
	EventTypeAuditApplicationCreate EventType = "audit.app.create"

	// EventTypeAuditApplicationDeleteRequest denotes an event where the CC
	// receives a request to delete an application.
	EventTypeAuditApplicationDeleteRequest EventType = "audit.app.delete-request"

	EventTypeAuditApplicationDropletMapped EventType = "audit.app.droplet.mapped"

	// EventTypeAuditApplicationMapRoute denotes an event where the CC maps an
	// application to a route.
	EventTypeAuditApplicationMapRoute EventType = "audit.app.map-route"

	// EventTypeAuditApplicationPackageCreate denotes an event where the CC creates an
	// application package.
	EventTypeAuditApplicationPackageCreate EventType = "audit.app.package.create"

	// EventTypeAuditApplicationPackageDelete denotes an event where the CC deletes an
	// application package.
	EventTypeAuditApplicationPackageDelete EventType = "audit.app.package.delete"

	// EventTypeAuditApplicationPackageDownload denotes an event where an application
	// package is downloaded.
	EventTypeAuditApplicationPackageDownload EventType = "audit.app.package.download"

	// EventTypeAuditApplicationPackageUpload denotes an event where an application
	// package is uploaded.
	EventTypeAuditApplicationPackageUpload EventType = "audit.app.package.upload"

	// EventTypeAuditApplicationRestage denotes an event where the CC restages an
	// application.
	EventTypeAuditApplicationRestage EventType = "audit.app.restage"

	EventTypeAuditApplicationSSHAuthorized EventType = "audit.app.ssh-authorized"

	EventTypeAuditApplicationSSHUnauthorized EventType = "audit.app.ssh-unauthorized"

	// EventTypeAuditApplicationStart denotes an event where the CC starts an
	// application.
	EventTypeAuditApplicationStart EventType = "audit.app.start"

	// EventTypeAuditApplicationStop denotes an event where the CC stops an
	// application.
	EventTypeAuditApplicationStop EventType = "audit.app.stop"

	// EventTypeAuditApplicationUnmapRoute denotes an event where the CC unmaps
	// an application from a route.
	EventTypeAuditApplicationUnmapRoute EventType = "audit.app.unmap-route"

	// EventTypeAuditApplicationUpdate denotes an event where the CC updates an
	// application.
	EventTypeAuditApplicationUpdate EventType = "audit.app.update"

	// EventTypeAuditApplicationUploadBits denotes an event where application
	// bits are uploaded to the CC.
	EventTypeAuditApplicationUploadBits EventType = "audit.app.upload-bits"

	// EventTypeOrganizationCreate denotes an event where the CC creates an
	// organization.
	EventTypeOrganizationCreate EventType = "audit.organization.create"

	// EventTypeOrganizationDeleteRequest denotes an event where the CC receives
	// request to delete an organization.
	EventTypeOrganizationDeleteRequest EventType = "audit.organization.delete-request"

	// EventTypeOrganizationUpdate denotes an event where the CC updates an
	// organization.
	EventTypeOrganizationUpdate EventType = "audit.organization.update"

	// EventTypeAuditRouteCreate denotes an event where the CC creates a route.
	EventTypeAuditRouteCreate EventType = "audit.route.create"

	// EventTypeAuditRouteDeleteRequest denotes an event where the CC receives a
	// request to delete a route.
	EventTypeAuditRouteDeleteRequest EventType = "audit.route.delete-request"

	// EventTypeAuditRouteUpdate denotes an event where the CC updates a route.
	EventTypeAuditRouteUpdate EventType = "audit.route.update"

	// EventTypeAuditServiceCreate denotes an event where the CC creates a service.
	EventTypeAuditServiceCreate EventType = "audit.service.create"

	// EventTypeAuditServiceDelete denotes an event where the CC deletes a service.
	EventTypeAuditServiceDelete EventType = "audit.service.delete"

	// EventTypeAuditServiceUpdate denotes an event where the CC updates a service.
	EventTypeAuditServiceUpdate EventType = "audit.service.update"

	// EventTypeAuditServiceBindingCreate denotes an event where the CC creates
	// a service binding.
	EventTypeAuditServiceBindingCreate EventType = "audit.service_binding.create"

	// EventTypeAuditServiceBindingDelete denotes an event where the CC deletes
	// a service binding.
	EventTypeAuditServiceBindingDelete EventType = "audit.service_binding.delete"

	// EventTypeAuditServiceBrokerCreate denotes an event where the CC creates
	// a service broker.
	EventTypeAuditServiceBrokerCreate EventType = "audit.service_broker.create"

	// EventTypeAuditServiceBrokerDelete denotes an event where the CC deletes
	// a service broker.
	EventTypeAuditServiceBrokerDelete EventType = "audit.service_broker.delete"

	// EventTypeAuditServiceBrokerUpdate denotes an event where the CC updates
	// a service broker.
	EventTypeAuditServiceBrokerUpdate EventType = "audit.service_broker.update"

	EventTypeAuditServiceDashboardClientCreate EventType = "audit.service_dashboard_client.create"

	EventTypeAuditServiceDashboardClientDelete EventType = "audit.service_dashboard_client.delete"

	EventTypeServiceInstanceBindRoute EventType = "audit.service_instance.bind_route"

	// EventTypeAuditServiceInstanceCreate denotes an event where the CC creates
	// a service instance.
	EventTypeAuditServiceInstanceCreate EventType = "audit.service_instance.create"

	// EventTypeAuditServiceInstanceDelete denotes an event where the CC deletes
	// a service instance.
	EventTypeAuditServiceInstanceDelete EventType = "audit.service_instance.delete"

	EventTypeServiceInstanceUnbindRoute EventType = "audit.service_instance.unbind_route"

	// EventTypeAuditServiceInstanceUpdate denotes an event where the CC updates
	// a service instance.
	EventTypeAuditServiceInstanceUpdate EventType = "audit.service_instance.update"

	EventTypeAuditServiceKeyCreate EventType = "audit.service_key.create"

	EventTypeAuditServiceKeyDelete EventType = "audit.service_key.delete"

	// EventTypeAuditServicePlanCreate denotes an event where the CC creates
	// a service plan.
	EventTypeAuditServicePlanCreate EventType = "audit.service_plan.create"

	// EventTypeAuditServicePlanDelete denotes an event where the CC deletes
	// a service plan.
	EventTypeAuditServicePlanDelete EventType = "audit.service_plan.delete"

	// EventTypeAuditServicePlanUpdate denotes an event where the CC updates
	// a service plan.
	EventTypeAuditServicePlanUpdate EventType = "audit.service_plan.update"

	EventTypeAuditServicePlanVisibilityCreate EventType = "audit.service_plan_visibility.create"

	EventTypeAuditServicePlanVisibilityDelete EventType = "audit.service_plan_visibility.delete"

	EventTypeAuditServicePlanVisibilityUpdate EventType = "audit.service_plan_visibility.update"

	// EventTypeAuditSpaceCreate denotes an event where the CC creates a space.
	EventTypeAuditSpaceCreate EventType = "audit.space.create"

	// EventTypeAuditSpaceDeleteRequest denotes an event where the CC receives a
	// request to delete a space.
	EventTypeAuditSpaceDeleteRequest EventType = "audit.space.delete-request"

	// EventTypeAuditSpaceUpdate denotes an event where the CC updates a space.
	EventTypeAuditSpaceUpdate EventType = "audit.space.update"

	// EventTypeAuditUserProvidedServiceInstanceCreate denotes an event where the
	// CC creates a user provided service instance.
	EventTypeAuditUserProvidedServiceInstanceCreate EventType = "audit.user_provided_service_instance.create"

	// EventTypeAuditUserProvidedServiceInstanceDelete denotes an event where the
	// CC deletes a user provided service instance.
	EventTypeAuditUserProvidedServiceInstanceDelete EventType = "audit.user_provided_service_instance.delete"

	// EventTypeAuditUserProvidedServiceInstanceUpdate denotes an event where the
	// CC updates a user provided service instance.
	EventTypeAuditUserProvidedServiceInstanceUpdate EventType = "audit.user_provided_service_instance.update"

	// EventTypeAuditUserSpaceAuditorAdd denotes an event where the CC associates
	// an auditor with a space.
	EventTypeAuditUserSpaceAuditorAdd EventType = "audit.user.space_auditor_add"

	// EventTypeAuditUserSpaceAuditorRemove denotes an event where the CC removes
	// an auditor from a space.
	EventTypeAuditUserSpaceAuditorRemove EventType = "audit.user.space_auditor_remove"

	// EventTypeAuditUserSpaceManagerAdd denotes an event where the CC associates
	// a manager with a space.
	EventTypeAuditUserSpaceManagerAdd EventType = "audit.user.space_manager_add"

	// EventTypeAuditUserSpaceManagerRemove denotes an event where the CC removes
	// a manager from a space.
	EventTypeAuditUserSpaceManagerRemove EventType = "audit.user.space_manager_remove"

	// EventTypeAuditUserSpaceDeveloperAdd denotes an event where the CC
	// associates a developer with a space.
	EventTypeAuditUserSpaceDeveloperAdd EventType = "audit.user.space_developer_add"

	// EventTypeAuditUserSpaceDeveloperRemove denotes an event where the CC removes
	// a developer from a space.
	EventTypeAuditUserSpaceDeveloperRemove EventType = "audit.user.space_developer_remove"

	// EventTypeAuditUserOrganizationAuditorAdd denotes an event where the CC
	// associates an auditor with an organization.
	EventTypeAuditUserOrganizationAuditorAdd EventType = "audit.user.organization_auditor_add"

	// EventTypeAuditUserOrganizationAuditorRemove denotes an event where the CC
	// removes an auditor from an organization.
	EventTypeAuditUserOrganizationAuditorRemove EventType = "audit.user.organization_auditor_remove"

	// EventTypeAuditUserOrganizationBillingManagerAdd denotes an event where the CC
	// associates a billing manager with an organization.
	EventTypeAuditUserOrganizationBillingManagerAdd EventType = "audit.user.organization_billing_manager_add"

	// EventTypeAuditUserOrganizationBillingManagerRemove denotes an event where the CC
	// removes a billing manager from an organization.
	EventTypeAuditUserOrganizationBillingManagerRemove EventType = "audit.user.organization_billing_manager_remove"

	// EventTypeAuditUserOrganizationManagerAdd denotes an event where the CC
	// associates a manager with an organization.
	EventTypeAuditUserOrganizationManagerAdd EventType = "audit.user.organization_manager_add"

	// EventTypeAuditUserOrganizationManagerRemove denotes an event where the CC
	// removes a manager from an organization.
	EventTypeAuditUserOrganizationManagerRemove EventType = "audit.user.organization_manager_remove"

	// EventTypeAuditUserOrganizationUserAdd denotes an event where the CC associates
	// an organization with a user.
	EventTypeAuditUserOrganizationUserAdd EventType = "audit.user.organization_user_add"

	// EventTypeAuditUserOrganizationUserRemove denotes an event where the CC
	// associates an organization with a user.
	EventTypeAuditUserOrganizationUserRemove EventType = "audit.user.organization_user_remove"

	EventTypeBlobRemoveOrphan EventType = "blob.remove_orphan"

	// "experimental" events

	// EventTypeAuditApplicationBuildCreate denotes an event where an application
	// build is created.
	EventTypeAuditApplicationBuildCreate EventType = "audit.app.build.create"

	// EventTypeAuditApplicationDropletCreate denotes an event where an application
	// droplet is created.
	EventTypeAuditApplicationDropletCreate EventType = "audit.app.droplet.create"

	// EventTypeAuditApplicationDropletDelete denotes an event where an application
	// droplet is deleted.
	EventTypeAuditApplicationDropletDelete EventType = "audit.app.droplet.delete"

	// EventTypeAuditApplicationDropletDownload denotes an event where an application
	// droplet is downloaded.
	EventTypeAuditApplicationDropletDownload EventType = "audit.app.droplet.download"

	// EventTypeAuditApplicationProcessCrash denotes an event where an application
	// process crashes.
	EventTypeAuditApplicationProcessCrash EventType = "audit.app.process.crash"

	// EventTypeAuditApplicationProcessCreate denotes an event where the CC
	// creates an application process.
	EventTypeAuditApplicationProcessCreate EventType = "audit.app.process.create"

	// EventTypeAuditApplicationProcessDelete denotes an event where the CC
	// deletes an application process.
	EventTypeAuditApplicationProcessDelete EventType = "audit.app.process.delete"

	// EventTypeAuditApplicationProcessScale denotes an event where the CC scales
	// an application process.
	EventTypeAuditApplicationProcessScale EventType = "audit.app.process.scale"

	// EventTypeAuditApplicationProcessTerminateInstance denotes an event where
	// the CC terminates an application process instance.
	EventTypeAuditApplicationProcessTerminateInstance EventType = "audit.app.process.terminate_instance"

	// EventTypeAuditApplicationProcessUpdate denotes an event where the CC
	// updates an application process.
	EventTypeAuditApplicationProcessUpdate EventType = "audit.app.process.update"

	// EventTypeAuditApplicationTaskCancel denotes an event where the CC cancels
	// a task.
	EventTypeAuditApplicationTaskCancel EventType = "audit.app.task.cancel"

	// EventTypeAuditApplicationTaskCreate denotes an event where the CC creates
	// a task.
	EventTypeAuditApplicationTaskCreate EventType = "audit.app.task.create"

	EventTypeServiceInstanceShare EventType = "audit.service_instance.share"

	EventTypeServiceInstanceUnshare EventType = "audit.service_instance.unshare"
)
