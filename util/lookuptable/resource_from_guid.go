package lookuptable

import "code.cloudfoundry.org/cli/v8/resources"

func OrgFromGUID(orgs []resources.Organization) map[string]resources.Organization {
	result := make(map[string]resources.Organization)
	for _, org := range orgs {
		result[org.GUID] = org
	}
	return result
}

func SpaceFromGUID(spaces []resources.Space) map[string]resources.Space {
	result := make(map[string]resources.Space)
	for _, space := range spaces {
		result[space.GUID] = space
	}
	return result
}

func AppFromGUID(apps []resources.Application) map[string]resources.Application {
	result := make(map[string]resources.Application)
	for _, app := range apps {
		result[app.GUID] = app
	}
	return result
}
