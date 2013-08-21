package api

type Authorizer interface {
	CanAccessOrg(userGuid string, orgName string) bool
}

type CloudControllerAuthorizer struct{}

func (a CloudControllerAuthorizer) CanAccessOrg(userGuid string, orgName string) bool {
	return true
}
