package v2action

func (actor Actor) CreateSharedDomain(domain string, routerGroup RouterGroup) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.CreateSharedDomain(domain, routerGroup.GUID)
	return Warnings(warnings), err
}
