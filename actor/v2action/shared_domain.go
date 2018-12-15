package v2action

func (actor Actor) CreateSharedDomain(domain string, routerGroup RouterGroup, isInternal bool) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateSharedDomain(domain, routerGroup.GUID, isInternal)
	return Warnings(warnings), err
}
