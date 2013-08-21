package api

type Authorizer interface {
	CanAccessOrg(userGuid string, orgName string) bool
	CanAccessSpace(userGuid string, spaceName string) bool
}

type CloudControllerAuthorizer struct{}

func (a CloudControllerAuthorizer) CanAccessOrg(userGuid string, orgName string) bool {
	return true
}

func (a CloudControllerAuthorizer) CanAccessSpace(userGuid string, spaceName string) bool {
	return true
}
